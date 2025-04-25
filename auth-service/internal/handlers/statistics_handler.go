package handlers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"time"

	"github.com/gofiber/fiber/v2"
)

func UserStatistics(c *fiber.Ctx) error {
	var stats serializers.UserStatsResponse
	now := time.Now().UTC()
	year := now.Year()
	monthStart := time.Date(year, now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

	// 1. Tổng số người dùng
	if err := db.DB.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error when counting users: "+err.Error()).Send(c)
	}

	// 2. Số người dùng mới trong tháng hiện tại
	if err := db.DB.Model(&models.User{}).
		Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
		Count(&stats.NewUsersInMonth).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error getting new users in month: "+err.Error()).Send(c)
	}

	// 3. Số người dùng theo tháng trong năm
	type monthResult struct {
		Month int
		Count int
	}
	var monthCounts []monthResult
	if err := db.DB.Model(&models.User{}).
		Select("EXTRACT(MONTH FROM created_at) as month, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", yearStart, yearEnd).
		Group("EXTRACT(MONTH FROM created_at)").
		Scan(&monthCounts).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error getting users each month: "+err.Error()).Send(c)
	}

	// Tạo mảng 12 tháng, điền số liệu từ truy vấn
	stats.UsersByMonth = make([]serializers.MonthlyUserCount, 12)
	for i := 1; i <= 12; i++ {
		stats.UsersByMonth[i-1] = serializers.MonthlyUserCount{Name: i, Count: 0}
		for _, mc := range monthCounts {
			if mc.Month == i {
				stats.UsersByMonth[i-1].Count = mc.Count
				break
			}
		}
	}
	// Chỉ lấy các tháng từ 1 đến tháng hiện tại
	stats.UsersByMonth = stats.UsersByMonth[:int(now.Month())]

	return responses.NewSuccessResponse(fiber.StatusOK, stats).Send(c)
}
