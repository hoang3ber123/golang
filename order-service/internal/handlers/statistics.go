package handlers

import (
	"fmt"
	"order-service/internal/db"
	"order-service/internal/responses"
	"order-service/internal/serializers"
	"order-service/internal/services"

	"github.com/gofiber/fiber/v2"
)

// API: Thống kê theo payment status theo ngày tháng năm
func GetOrderPaymentStatistic(c *fiber.Ctx) error {
	query := serializers.StatisticsQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}

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
        IFNULL(SUM(o.amount_paid), 0) AS total_amount
    FROM time_series ts
    CROSS JOIN payment_statuses ps
    LEFT JOIN orders o 
        ON DATE_FORMAT(o.created_at, '%s') = DATE_FORMAT(ts.time_unit, '%s') 
        AND o.payment_status = ps.payment_status
    GROUP BY time_unit, ps.payment_status
    ORDER BY ps.payment_status DESC,time_unit ASC;
`, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, timeFormat)

	type QueryResult struct {
		TimeUnit      string  `json:"time_unit"`
		PaymentStatus string  `json:"payment_status"`
		TotalAmount   float64 `json:"total_amount"`
	}
	var results []QueryResult
	if err := db.DB.Raw(queryStr).Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	fmt.Println("results: ", results)
	// tạo mảng response
	type ValuesResponse struct {
		TimeUnit    string  `json:"time_unit"`
		TotalAmount float64 `json:"total_amount"`
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
			TimeUnit:    results[rowIndex].TimeUnit,
			TotalAmount: results[rowIndex].TotalAmount,
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

// API: Thống kê tổng quát product,user,order theo payment status theo ngày tháng năm
func GetOrderGeneralStatistic(c *fiber.Ctx) error {
	query := serializers.StatisticsQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}

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
		IFNULL(SUM(o.amount_paid), 0) AS total_amount, 
		COUNT(DISTINCT o.user_id) AS total_users,
		COUNT(DISTINCT o.id) AS total_orders,
		COUNT(DISTINCT od.id) AS total_products
    FROM time_series ts
    CROSS JOIN payment_statuses ps
    LEFT JOIN orders o 
        ON DATE_FORMAT(o.created_at, '%s') = DATE_FORMAT(ts.time_unit, '%s') 
        AND o.payment_status = ps.payment_status
	LEFT JOIN order_details od
	    ON o.id = od.order_id
    GROUP BY time_unit, ps.payment_status
    ORDER BY ps.payment_status DESC,time_unit ASC;
`, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, timeFormat)

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

// API: Thống kê tổng quát product theo payment status theo ngày tháng năm
func GetOrderRankingProductStatistic(c *fiber.Ctx) error {
	query := serializers.StatisticsRankingQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}
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
sales_data AS (
    SELECT 
        DATE_FORMAT(o.created_at, '%s') AS time_unit,
        od.related_id,
        od.related_type,
        COUNT(DISTINCT o.id) AS total_orders, -- Tổng số hóa đơn chứa sản phẩm
        SUM(od.total_price) AS total_revenue -- Tổng doanh thu
    FROM orders o
    JOIN order_details od ON od.order_id = o.id
    WHERE o.payment_status = 'success' -- Chỉ lấy đơn hàng thành công
    GROUP BY time_unit, od.related_id, od.related_type
),
ranked_sales AS (
    SELECT 
        sd.*,
        RANK() OVER (PARTITION BY sd.time_unit ORDER BY sd.total_orders DESC) AS rank_position
    FROM sales_data sd
)
SELECT 
    ts.time_unit, 
    rs.related_id AS product_id, 
    rs.related_type AS product_type,
    rs.total_orders,
    rs.total_revenue,
    rs.rank_position
FROM time_series ts
LEFT JOIN ranked_sales rs 
    ON rs.time_unit = DATE_FORMAT(ts.time_unit, '%s') 
WHERE rs.rank_position <= %d
ORDER BY ts.time_unit ASC, rs.rank_position ASC;
`, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, query.TopRanking)

	type QueryResult struct {
		TimeUnit     string  `json:"time_unit"`
		ProductID    string  `json:"product_id"`
		ProductType  string  `json:"product_type"`
		TotalOrders  int     `json:"total_orders"`
		TotalRevenue float64 `json:"total_revenue"`
		RankPosition int     `json:"rank_position"`
	}
	var results []QueryResult
	if err := db.DB.Raw(queryStr).Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	fmt.Println("results: ", results)
	// tạo mảng response
	type ValuesResponse struct {
		TimeUnit     string  `json:"time_unit"`
		ProductID    string  `json:"product_id"`
		ProductType  string  `json:"product_type"`
		TotalOrders  int     `json:"total_orders"`
		TotalRevenue float64 `json:"total_revenue"`
		RankPosition int     `json:"rank_position"`
	}
	lenResult := len(results)
	response := make([]ValuesResponse, lenResult)
	// grouping dữ liệu vào response theo payment_status
	for rowIndex := range results {
		response[rowIndex] = ValuesResponse{
			TimeUnit:     results[rowIndex].TimeUnit,
			ProductID:    results[rowIndex].ProductID,
			ProductType:  results[rowIndex].ProductType,
			TotalOrders:  results[rowIndex].TotalOrders,
			TotalRevenue: results[rowIndex].TotalRevenue,
			RankPosition: results[rowIndex].RankPosition,
		}
	}
	totalResponse := serializers.StatisticsRanking{
		StatisticsRankingQuerySerializer: query,
		Chart:                            response,
	}
	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
}
