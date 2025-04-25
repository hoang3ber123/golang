package handlers

import (
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ProductStatistics(c *fiber.Ctx) error {

	var stats serializers.ProductStatsResponse
	now := time.Now().UTC()
	year := now.Year()
	monthStart := time.Date(year, now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

	// 1. Tổng số sản phẩm
	if err := db.DB.Model(&models.Product{}).Count(&stats.TotalProducts).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error when counting product:"+err.Error()).Send(c)
	}

	// 2. Số sản phẩm mới trong tháng hiện tại
	if err := db.DB.Model(&models.Product{}).
		Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
		Count(&stats.NewProductsInMonth).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error get new product in month:"+err.Error()).Send(c)
	}

	// 3. Số sản phẩm theo tháng trong năm
	type monthResult struct {
		Month int
		Count int
	}
	var monthCounts []monthResult
	if err := db.DB.Model(&models.Product{}).
		Select("EXTRACT(MONTH FROM created_at) as month, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", yearStart, yearEnd).
		Group("EXTRACT(MONTH FROM created_at)").
		Scan(&monthCounts).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error get product each month:"+err.Error()).Send(c)
	}

	// Tạo mảng 12 tháng, điền số liệu từ truy vấn
	stats.ProductsByMonth = make([]serializers.MonthlyProductCount, 12)
	for i := 1; i <= 12; i++ {
		stats.ProductsByMonth[i-1] = serializers.MonthlyProductCount{Name: i, Count: 0}
		for _, mc := range monthCounts {
			if mc.Month == i {
				stats.ProductsByMonth[i-1].Count = mc.Count
				break
			}
		}
	}
	// Chỉ lấy các tháng từ 1 đến tháng hiện tại
	stats.ProductsByMonth = stats.ProductsByMonth[:int(now.Month())]

	// 4. Số sản phẩm theo danh mục
	var categoryCounts []struct {
		Title string
		Count int
	}
	if err := db.DB.Model(&models.Category{}).
		Select("categories.title, COUNT(product_categories.product_id) as count").
		Joins("LEFT JOIN product_categories ON categories.id = product_categories.category_id").
		Group("categories.title").
		Scan(&categoryCounts).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error get product each category:"+err.Error()).Send(c)
	}

	// Chuyển đổi sang format response
	stats.ProductsByCategory = make([]serializers.CategoryProductCount, len(categoryCounts))
	for i, cc := range categoryCounts {
		stats.ProductsByCategory[i] = serializers.CategoryProductCount{
			Name:  cc.Title,
			Count: cc.Count,
		}
	}

	return responses.NewSuccessResponse(fiber.StatusOK, stats).Send(c)
}
