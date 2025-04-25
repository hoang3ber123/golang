package handlers

import (
	"fmt"
	"order-service/internal/db"
	"order-service/internal/models"
	"order-service/internal/responses"
	"order-service/internal/serializers"
	"order-service/internal/services"
	"order-service/pagination"

	"github.com/gofiber/fiber/v2"
)

func OrderList(c *fiber.Ctx) error {
	// lấy user từ context
	user := c.Locals("user").(*models.User)
	// Initializer query
	query := db.DB.Model(&models.Order{}).Where("user_id = ?", user.ID).Order("created_at DESC")
	// Sử dụng hàm phân trang
	var instance []models.Order
	paginator, err := pagination.PaginateWithGORM(c, query, &instance)
	if err != nil {
		return err.Send(c)
	}
	var result interface{}
	if instance != nil {
		result = serializers.OrderListResponse(&instance)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
	}).Send(c)
}

func OrderDetail(c *fiber.Ctx) error {
	// lấy user từ context
	user := c.Locals("user").(*models.User)
	id := c.Params("id")
	// Sử dụng hàm phân trang
	var instance models.Order
	db.DB.Model(&models.Order{}).Where("user_id = ? AND id = ?", user.ID, id).Find(&instance)
	// serializer
	result, err := serializers.OrderDetailResponse(&instance)
	if err != nil {
		return err.Send(c)
	}
	return responses.NewSuccessResponse(fiber.StatusOK, result).Send(c)
}

// API: Thống kê tổng quát product,user,order theo payment status theo ngày tháng năm
func OrderStatistic(c *fiber.Ctx) error {
	query := serializers.StatisticsQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}
	// lấy user
	user := c.Locals("user").(*models.User)
	var timeFormat, startTemp, stopCondition, timeIncrement string
	switch query.DayUnit {
	case "day":
		timeFormat = "%Y-%m-%d"
		startTemp = query.StartDay
		stopCondition = query.EndDay
		timeIncrement = "1 DAY"
	case "month":
		timeFormat = "%Y-%m"
		startTemp, _ = services.StartOfMonth(query.StartDay)
		stopCondition, _ = services.StartOfMonth(query.EndDay)
		timeIncrement = "1 MONTH"
	case "year":
		timeFormat = "%Y"
		startTemp, _ = services.StartOfYear(query.StartDay)
		stopCondition, _ = services.StartOfYear(query.EndDay)
		timeIncrement = "1 YEAR"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid day_unit"})
	}

	queryStr := fmt.Sprintf(`
		WITH RECURSIVE time_series AS (
			SELECT '%s' AS time_unit
			UNION ALL
			SELECT DATE_ADD(time_unit, INTERVAL %s)
			FROM time_series
			WHERE time_unit < '%s'
		),
		payment_statuses AS (
        SELECT 'success' AS payment_status
        UNION ALL
        SELECT 'cancel'
		)
		SELECT 
    DATE_FORMAT(ts.time_unit, '%s') AS time_unit, 
    ps.payment_status, 
    IFNULL(SUM(o.amount_paid), 0) AS total_amount, -- Tổng tiền của các hóa đơn
    COUNT(DISTINCT o.id) AS total_orders, -- Số lượng hóa đơn
    COUNT(od.id) AS total_order_details -- Số lượng order_details
FROM time_series ts
CROSS JOIN payment_statuses ps
LEFT JOIN orders o 
    ON DATE_FORMAT(o.created_at, '%s') = DATE_FORMAT(ts.time_unit, '%s') 
    AND o.payment_status = ps.payment_status
LEFT JOIN order_details od
    ON o.id = od.order_id
WHERE o.user_id = %s
GROUP BY time_unit, ps.payment_status
ORDER BY ps.payment_status DESC, time_unit ASC;
`, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, timeFormat, user.ID.String())

	type QueryResult struct {
		TimeUnit      string `json:"time_unit"`
		PaymentStatus string `json:"payment_status"`
		TotalUsers    int    `json:"total_users"`
		TotalOrders   int    `json:"total_orders"`
		TotalProducts int    `json:"total_products"`
	}
	var results []QueryResult
	if err := db.DB.Raw(queryStr).Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	fmt.Println("results: ", results)
	// tạo mảng response
	type ValuesResponse struct {
		TimeUnit      string `json:"time_unit"`
		TotalUsers    int    `json:"total_users"`
		TotalOrders   int    `json:"total_orders"`
		TotalProducts int    `json:"total_products"`
	}
	countValuesInTimeUnit := services.CountTimeUnits(query.StartDay, query.EndDay, query.DayUnit)
	response := map[string][]ValuesResponse{
		"success": make([]ValuesResponse, countValuesInTimeUnit),
		"cancel":  make([]ValuesResponse, countValuesInTimeUnit),
	}
	// grouping dữ liệu vào response theo payment_status
	indexTimeUnit := 0
	for rowIndex := range results {
		response[results[rowIndex].PaymentStatus][indexTimeUnit] = ValuesResponse{
			TimeUnit:      results[rowIndex].TimeUnit,
			TotalOrders:   results[rowIndex].TotalOrders,
			TotalProducts: results[rowIndex].TotalProducts,
			TotalUsers:    results[rowIndex].TotalUsers,
		}
		indexTimeUnit++
		if (indexTimeUnit) == countValuesInTimeUnit {
			indexTimeUnit = 0
		}
	}
	totalResponse := serializers.Statistics{
		StatisticsQuerySerializer: query,
		Chart:                     response,
	}
	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
}
