package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"order-service/internal/db"
	"order-service/internal/models"
	"order-service/internal/responses"
	"order-service/internal/serializers"
	"order-service/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetUserOrderStatistic thống kê đơn hàng theo payment_status cho user cụ thể
func GetUserOrderStatistic(c *fiber.Ctx) error {
	// Lấy user từ context
	user, _ := c.Locals("user").(*models.User)

	// Validate query
	query := serializers.StatisticsQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Xác định format thời gian và khoảng thời gian
	var timeFormat, startDay, endDay string
	switch query.DayUnit {
	case "day":
		timeFormat = "%Y-%m-%d"
		startDay = query.StartDay
		endDay = query.EndDay
	case "month":
		timeFormat = "%Y-%m"
		startDay, _ = services.StartOfMonth(query.StartDay)
		endDay, _ = services.EndOfMonth(query.EndDay)
	case "year":
		timeFormat = "%Y"
		startDay, _ = services.StartOfYear(query.StartDay)
		endDay, _ = services.EndOfYear(query.EndDay)
	default:
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid day_unit").Send(c)
	}

	// Struct cho kết quả truy vấn
	type QueryResult struct {
		TimeUnit      string  `json:"time_unit"`
		PaymentStatus string  `json:"payment_status"`
		TotalAmount   float64 `json:"total_amount"`
	}
	var results []QueryResult
	if err := db.DB.Model(&models.Order{}).
		Select(
			fmt.Sprintf("DATE_FORMAT(created_at, '%s') as time_unit", timeFormat),
			"payment_status",
			"COALESCE(SUM(amount_paid), 0) as total_amount",
		).
		Where("created_at BETWEEN ? AND ?", startDay, endDay).
		Where("user_id = ?", user.ID).
		Group("time_unit, payment_status").
		Order("payment_status DESC, time_unit ASC").
		Scan(&results).Error; err != nil {
		log.Printf("Error querying user payment statistics: %v", err)
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error querying user payment statistics: "+err.Error()).Send(c)
	}

	// Tạo response
	type ValuesResponse struct {
		TimeUnit    string  `json:"time_unit"`
		TotalAmount float64 `json:"total_amount"`
	}
	response := map[string][]ValuesResponse{}
	// Grouping dữ liệu vào response
	for _, result := range results {
		response[result.PaymentStatus] = append(response[result.PaymentStatus], ValuesResponse{
			TimeUnit:    result.TimeUnit,
			TotalAmount: result.TotalAmount,
		})
	}

	// Tạo response cuối cùng
	totalResponse := serializers.Statistics{
		StatisticsQuerySerializer: query,
		Chart:                     response,
	}
	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
}

// API: Thống kê theo payment status theo ngày tháng năm
func GetOrderPaymentStatistic(c *fiber.Ctx) error {
	query := serializers.StatisticsQuerySerializer{}
	if err := query.IsValid(c); err != nil {
		return err.Send(c)
	}

	var timeFormat, startDay, endDay string
	switch query.DayUnit {
	case "day":
		timeFormat = "%Y-%m-%d"
		startDay = query.StartDay
		endDay = query.EndDay
	case "month":
		timeFormat = "%Y-%m"
		startDay, _ = services.StartOfMonth(query.StartDay)
		endDay, _ = services.EndOfMonth(query.EndDay)
	case "year":
		timeFormat = "%Y"
		startDay, _ = services.StartOfYear(query.StartDay)
		endDay, _ = services.EndOfYear(query.EndDay)
	default:
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid day_unit").Send(c)
	}
	type QueryResult struct {
		TimeUnit      string  `json:"time_unit"`
		PaymentStatus string  `json:"payment_status"`
		TotalAmount   float64 `json:"total_amount"`
	}
	var results []QueryResult
	if err := db.DB.Model(&models.Order{}).
		Select(
			fmt.Sprintf("DATE_FORMAT(created_at, '%s') as time_unit", timeFormat),
			"payment_status",
			"COALESCE(SUM(amount_paid), 0) as total_amount",
		).
		Where("created_at BETWEEN ? AND ?", startDay, endDay).
		Group("time_unit, payment_status").
		Order("payment_status DESC,time_unit ASC").
		Scan(&results).Error; err != nil {
		log.Printf("Error querying payment statistics: %v", err)
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error querying payment statistics: "+err.Error()).Send(c)
	}
	// tạo mảng response
	type ValuesResponse struct {
		TimeUnit    string  `json:"time_unit"`
		TotalAmount float64 `json:"total_amount"`
	}
	response := map[string][]ValuesResponse{}
	// duyệt mảng results grouping dữ liệu vào response theo payment_status
	for index := range results {
		response[results[index].PaymentStatus] = append(response[results[index].PaymentStatus], ValuesResponse{
			TimeUnit:    results[index].TimeUnit,
			TotalAmount: results[index].TotalAmount,
		})
	}

	totalResponse := serializers.Statistics{
		StatisticsQuerySerializer: query,
		Chart:                     response,
	}
	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
}

// Hàm API thống kê Payment
func PaymentStatistics(c *fiber.Ctx) error {
	var stats serializers.PaymentStatsResponse
	now := time.Now().UTC()
	year := now.Year()
	monthStart := time.Date(year, now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

	// 1. Tổng AmountPaid
	var totalAmountPaid sql.NullFloat64
	if err := db.DB.Model(&models.Order{}).
		Select("SUM(amount_paid)").
		Scan(&totalAmountPaid).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error calculating total amount paid: "+err.Error()).Send(c)
	}
	stats.TotalAmountPaid = totalAmountPaid.Float64

	// 2. AmountPaid mới trong tháng hiện tại
	var newAmountPaid sql.NullFloat64
	if err := db.DB.Model(&models.Order{}).
		Select("SUM(amount_paid)").
		Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
		Scan(&newAmountPaid).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error calculating new amount paid in month: "+err.Error()).Send(c)
	}
	stats.NewAmountPaidInMonth = newAmountPaid.Float64

	// 3. AmountPaid theo tháng trong năm
	type monthResult struct {
		Month  int
		Amount float64
	}
	var monthAmounts []monthResult
	if err := db.DB.Model(&models.Order{}).
		Select("EXTRACT(MONTH FROM created_at) as month, SUM(amount_paid) as amount").
		Where("created_at BETWEEN ? AND ?", yearStart, yearEnd).
		Group("EXTRACT(MONTH FROM created_at)").
		Scan(&monthAmounts).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error calculating amount paid by month: "+err.Error()).Send(c)
	}

	// Tạo mảng 12 tháng, điền số liệu từ truy vấn
	stats.AmountPaidByMonth = make([]serializers.MonthlyAmountPaid, 12)
	for i := 1; i <= 12; i++ {
		stats.AmountPaidByMonth[i-1] = serializers.MonthlyAmountPaid{Name: i, Amount: 0}
		for _, ma := range monthAmounts {
			if ma.Month == i {
				stats.AmountPaidByMonth[i-1].Amount = ma.Amount
				break
			}
		}
	}
	// Chỉ lấy các tháng từ 1 đến tháng hiện tại
	stats.AmountPaidByMonth = stats.AmountPaidByMonth[:int(now.Month())]

	// 4. AmountPaid theo PaymentStatus
	var statusAmounts []struct {
		PaymentStatus string  `gorm:"column:payment_status"`
		Amount        float64 `gorm:"column:amount"`
	}
	if err := db.DB.Model(&models.Order{}).
		Select("payment_status, SUM(amount_paid) as amount").
		Group("payment_status").
		Scan(&statusAmounts).Error; err != nil {
		log.Printf("Error querying payment status: %v", err)
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error calculating amount paid by status: "+err.Error()).Send(c)
	}

	// Log dữ liệu thô để kiểm tra
	log.Printf("Raw status amounts: %+v", statusAmounts)

	// Tạo map để ánh xạ status với amount
	statusAmountMap := make(map[string]float64)
	for _, sa := range statusAmounts {
		statusAmountMap[sa.PaymentStatus] = sa.Amount
	}

	// Tính phần trăm và tạo response
	for status := range models.AllowPaymentStatus {
		amount := statusAmountMap[status]
		percentage := 0.0
		if stats.TotalAmountPaid > 0 {
			percentage = (amount / stats.TotalAmountPaid) * 100
			// Làm tròn 2 chữ số thập phân
			percentage = math.Round(percentage*100) / 100
		}
		stats.AmountPaidByStatus = append(stats.AmountPaidByStatus, serializers.StatusAmountPaid{
			Status:     status,
			Percentage: percentage,
		})
	}

	return responses.NewSuccessResponse(fiber.StatusOK, stats).Send(c)
}

// API: Thống kê tổng quát product,user,order theo payment status theo ngày tháng năm
// func GetOrderGeneralStatistic(c *fiber.Ctx) error {
// 	query := serializers.StatisticsQuerySerializer{}
// 	if err := query.IsValid(c); err != nil {
// 		return err.Send(c)
// 	}

// 	var timeFormat, startTemp, stopCondition, timeIncrement string
// 	switch query.DayUnit {
// 	case "day":
// 		timeFormat = "%Y-%m-%d"
// 		startTemp = query.StartDay
// 		stopCondition = query.EndDay
// 		timeIncrement = "1 DAY"
// 	case "month":
// 		timeFormat = "%Y-%m"
// 		startTemp, _ = services.StartOfMonth(query.StartDay)
// 		stopCondition, _ = services.StartOfMonth(query.EndDay)
// 		timeIncrement = "1 MONTH"
// 	case "year":
// 		timeFormat = "%Y"
// 		startTemp, _ = services.StartOfYear(query.StartDay)
// 		stopCondition, _ = services.StartOfYear(query.EndDay)
// 		timeIncrement = "1 YEAR"
// 	default:
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid day_unit"})
// 	}

// 	queryStr := fmt.Sprintf(`
// 		WITH RECURSIVE time_series AS (
// 			SELECT '%s' AS time_unit
// 			UNION ALL
// 			SELECT DATE_ADD(time_unit, INTERVAL %s)
// 			FROM time_series
// 			WHERE time_unit < '%s'
// 		),
// 		payment_statuses AS (
//         SELECT 'success' AS payment_status
//         UNION ALL
//         SELECT 'cancel'
// 		)
// 		SELECT
//         DATE_FORMAT(ts.time_unit, '%s') AS time_unit,
// 		ps.payment_status,
// 		IFNULL(SUM(o.amount_paid), 0) AS total_amount,
// 		COUNT(DISTINCT o.user_id) AS total_users,
// 		COUNT(DISTINCT o.id) AS total_orders,
// 		COUNT(DISTINCT od.id) AS total_products
//     FROM time_series ts
//     CROSS JOIN payment_statuses ps
//     LEFT JOIN orders o
//         ON DATE_FORMAT(o.created_at, '%s') = DATE_FORMAT(ts.time_unit, '%s')
//         AND o.payment_status = ps.payment_status
// 	LEFT JOIN order_details od
// 	    ON o.id = od.order_id
//     GROUP BY time_unit, ps.payment_status
//     ORDER BY ps.payment_status DESC,time_unit ASC;
// `, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, timeFormat)

// 	type QueryResult struct {
// 		TimeUnit      string `json:"time_unit"`
// 		PaymentStatus string `json:"payment_status"`
// 		TotalUsers    int    `json:"total_users"`
// 		TotalOrders   int    `json:"total_orders"`
// 		TotalProducts int    `json:"total_products"`
// 	}
// 	var results []QueryResult
// 	if err := db.DB.Raw(queryStr).Scan(&results).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	fmt.Println("results: ", results)
// 	// tạo mảng response
// 	type ValuesResponse struct {
// 		TimeUnit      string `json:"time_unit"`
// 		TotalUsers    int    `json:"total_users"`
// 		TotalOrders   int    `json:"total_orders"`
// 		TotalProducts int    `json:"total_products"`
// 	}
// 	countValuesInTimeUnit := services.CountTimeUnits(query.StartDay, query.EndDay, query.DayUnit)
// 	response := map[string][]ValuesResponse{
// 		"success": make([]ValuesResponse, countValuesInTimeUnit),
// 		"cancel":  make([]ValuesResponse, countValuesInTimeUnit),
// 	}
// 	// grouping dữ liệu vào response theo payment_status
// 	indexTimeUnit := 0
// 	for rowIndex := range results {
// 		response[results[rowIndex].PaymentStatus][indexTimeUnit] = ValuesResponse{
// 			TimeUnit:      results[rowIndex].TimeUnit,
// 			TotalOrders:   results[rowIndex].TotalOrders,
// 			TotalProducts: results[rowIndex].TotalProducts,
// 			TotalUsers:    results[rowIndex].TotalUsers,
// 		}
// 		indexTimeUnit++
// 		if (indexTimeUnit) == countValuesInTimeUnit {
// 			indexTimeUnit = 0
// 		}
// 	}
// 	totalResponse := serializers.Statistics{
// 		StatisticsQuerySerializer: query,
// 		Chart:                     response,
// 	}
// 	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
// }

// API: Thống kê tổng quát product theo payment status theo ngày tháng năm
// func GetOrderRankingProductStatistic(c *fiber.Ctx) error {
// 	query := serializers.StatisticsRankingQuerySerializer{}
// 	if err := query.IsValid(c); err != nil {
// 		return err.Send(c)
// 	}
// 	var timeFormat, startTemp, stopCondition, timeIncrement string
// 	switch query.DayUnit {
// 	case "day":
// 		timeFormat = "%Y-%m-%d"
// 		startTemp = query.StartDay
// 		stopCondition = query.EndDay
// 		timeIncrement = "1 DAY"
// 	case "month":
// 		timeFormat = "%Y-%m"
// 		startTemp, _ = services.StartOfMonth(query.StartDay)
// 		stopCondition, _ = services.StartOfMonth(query.EndDay)
// 		timeIncrement = "1 MONTH"
// 	case "year":
// 		timeFormat = "%Y"
// 		startTemp, _ = services.StartOfYear(query.StartDay)
// 		stopCondition, _ = services.StartOfYear(query.EndDay)
// 		timeIncrement = "1 YEAR"
// 	default:
// 		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid day_unit").Send(c)
// 	}

// 	queryStr := fmt.Sprintf(`
// 		WITH RECURSIVE time_series AS (
// 			SELECT '%s' AS time_unit
// 			UNION ALL
// 			SELECT DATE_ADD(time_unit, INTERVAL %s)
// 			FROM time_series
// 			WHERE time_unit < '%s'
// 		),
// sales_data AS (
//     SELECT
//         DATE_FORMAT(o.created_at, '%s') AS time_unit,
//         od.related_id,
//         od.related_type,
//         COUNT(DISTINCT o.id) AS total_orders, -- Tổng số hóa đơn chứa sản phẩm
//         SUM(od.total_price) AS total_revenue -- Tổng doanh thu
//     FROM orders o
//     JOIN order_details od ON od.order_id = o.id
//     WHERE o.payment_status = 'success' -- Chỉ lấy đơn hàng thành công
//     GROUP BY time_unit, od.related_id, od.related_type
// ),
// ranked_sales AS (
//     SELECT
//         sd.*,
//         RANK() OVER (PARTITION BY sd.time_unit ORDER BY sd.total_orders DESC) AS rank_position
//     FROM sales_data sd
// )
// SELECT
//     ts.time_unit,
//     rs.related_id AS product_id,
//     rs.related_type AS product_type,
//     rs.total_orders,
//     rs.total_revenue,
//     rs.rank_position
// FROM time_series ts
// LEFT JOIN ranked_sales rs
//     ON rs.time_unit = DATE_FORMAT(ts.time_unit, '%s')
// WHERE rs.rank_position <= %d
// ORDER BY ts.time_unit ASC, rs.rank_position ASC;
// `, startTemp, timeIncrement, stopCondition, timeFormat, timeFormat, query.TopRanking)

// 	type QueryResult struct {
// 		TimeUnit     string  `json:"time_unit"`
// 		ProductID    string  `json:"product_id"`
// 		ProductType  string  `json:"product_type"`
// 		TotalOrders  int     `json:"total_orders"`
// 		TotalRevenue float64 `json:"total_revenue"`
// 		RankPosition int     `json:"rank_position"`
// 	}
// 	var results []QueryResult
// 	if err := db.DB.Raw(queryStr).Scan(&results).Error; err != nil {
// 		return responses.NewErrorResponse(fiber.StatusInternalServerError, "error"+err.Error()).Send(c)
// 	}
// 	fmt.Println("results: ", results)
// 	// tạo mảng response
// 	type ValuesResponse struct {
// 		TimeUnit     string  `json:"time_unit"`
// 		ProductID    string  `json:"product_id"`
// 		ProductType  string  `json:"product_type"`
// 		TotalOrders  int     `json:"total_orders"`
// 		TotalRevenue float64 `json:"total_revenue"`
// 		RankPosition int     `json:"rank_position"`
// 	}
// 	lenResult := len(results)
// 	response := make([]ValuesResponse, lenResult)
// 	// grouping dữ liệu vào response theo payment_status
// 	for rowIndex := range results {
// 		response[rowIndex] = ValuesResponse{
// 			TimeUnit:     results[rowIndex].TimeUnit,
// 			ProductID:    results[rowIndex].ProductID,
// 			ProductType:  results[rowIndex].ProductType,
// 			TotalOrders:  results[rowIndex].TotalOrders,
// 			TotalRevenue: results[rowIndex].TotalRevenue,
// 			RankPosition: results[rowIndex].RankPosition,
// 		}
// 	}
// 	totalResponse := serializers.StatisticsRanking{
// 		StatisticsRankingQuerySerializer: query,
// 		Chart:                            response,
// 	}
// 	return responses.NewSuccessResponse(fiber.StatusOK, totalResponse).Send(c)
// }
