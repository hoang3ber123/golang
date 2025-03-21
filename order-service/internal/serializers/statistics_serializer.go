package serializers

import (
	"order-service/internal/responses"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

// ProductQuerySerializer định nghĩa các tham số truy vấn để lọc và sắp xếp product
type StatisticsQuerySerializer struct {
	StartDay string `query:"start_day" json:"start_day" validate:"required"` // Định dạng YYYY-MM-DD
	EndDay   string `query:"end_day" json:"end_day" validate:"required"`     // Định dạng YYYY-MM-DD
	DayUnit  string `query:"day_unit" json:"day_unit" validate:"required,oneof=month day year"`
}

func (s *StatisticsQuerySerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.QueryParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid pagination parameters: "+err.Error())
	}
	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	// Validation tùy chỉnh cho ngày tháng
	if s.StartDay != "" {
		startTime, err := time.Parse("2006-01-02", s.StartDay)
		if err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid start_payment_day format: must be YYYY-MM-DD")
		}
		if s.EndDay != "" {
			endTime, err := time.Parse("2006-01-02", s.EndDay)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_payment_day format: must be YYYY-MM-DD")
			}
			if startTime.After(endTime) {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "start_payment_day must not be after end_payment_day")
			}
		}
	} else if s.EndDay != "" {
		if _, err := time.Parse("2006-01-02", s.EndDay); err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_payment_day format: must be YYYY-MM-DD")
		}
	}
	return nil
}

// StatisticsResponse là struct chính cho phản hồi API
type Statistics struct {
	StatisticsQuerySerializer
	Chart any `json:"chart"`
}

type StatisticsRankingQuerySerializer struct {
	StatisticsQuerySerializer
	TopRanking int `query:"top_ranking" json:"top_ranking" validate:"required"`
}

// StatisticsRankingResponse là struct chính cho phản hồi API về xếp hạng
type StatisticsRanking struct {
	StatisticsRankingQuerySerializer
	Chart any `json:"chart"`
}

func (s *StatisticsRankingQuerySerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse query params vào struct
	if err := c.QueryParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid query parameters: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	if s.TopRanking <= 0 {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "top_ranking must be greater than 0")
	}
	// Validation tùy chỉnh cho ngày tháng
	if s.StartDay != "" {
		startTime, err := time.Parse("2006-01-02", s.StartDay)
		if err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid start_day format: must be YYYY-MM-DD")
		}
		if s.EndDay != "" {
			endTime, err := time.Parse("2006-01-02", s.EndDay)
			if err != nil {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_day format: must be YYYY-MM-DD")
			}
			if startTime.After(endTime) {
				return responses.NewErrorResponse(fiber.StatusBadRequest, "start_day must not be after end_day")
			}
		}
	} else if s.EndDay != "" {
		if _, err := time.Parse("2006-01-02", s.EndDay); err != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid end_day format: must be YYYY-MM-DD")
		}
	}

	return nil
}
