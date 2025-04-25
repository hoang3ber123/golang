package serializers

import (
	"fmt"
	"product-service/internal/models"
	"product-service/internal/responses"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TaskCreateSerializer struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Mode        string `json:"mode" validate:"required,oneof=infinity minutes"`
	Frequency   int    `json:"frequency" validate:"required,gt=0"`
	StartedAt   string `json:"started_at"`
	StoppedAt   string `json:"stopped_at"`
}

func (s *TaskCreateSerializer) IsValid(c *fiber.Ctx) *responses.ErrorResponse {
	// Parse body to struct
	if err := c.BodyParser(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid input: "+err.Error())
	}

	// Basic validation với go-playground/validator
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Validate mode thủ công nếu cần so với models
	if s.Mode != "minutes" && s.Mode != "infinity" {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Mode must be 'minutes' or 'infinity'")
	}

	// Parse thời gian để kiểm tra logic
	layout := time.RFC3339 // format "2006-01-02T15:04:05"
	fmt.Println(time.Now())
	start, errStart := time.Parse(layout, s.StartedAt)
	now := time.Now()
	fmt.Println(start)
	// Nếu có lỗi parse thời gian, trả về lỗi
	if errStart != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid StartedAt datetime format")
	}

	// Nếu thời gian bắt đầu < thời gian hiện tại
	if !start.After(now) {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "StartedAt must be in the future")
	}

	// Nếu mode là "minutes" thì kiểm tra StoppedAt
	if s.Mode == "minutes" {
		if s.StoppedAt == "" {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "StoppedAt is required when mode is 'minutes'")
		}

		stop, errStop := time.Parse(layout, s.StoppedAt)

		// Kiểm tra lỗi parse StoppedAt
		if errStop != nil {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid StoppedAt datetime format")
		}

		// Kiểm tra StoppedAt phải sau StartedAt
		if !stop.After(start) {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "StoppedAt must be after StartedAt")
		}
	}

	// Nếu mode là "infinity", StoppedAt không cần thiết, không kiểm tra
	if s.Mode == "infinity" {
		if s.StoppedAt != "" {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "StoppedAt should not be provided when mode is 'infinity'")
		}
	}

	// Nếu không có lỗi, trả về nil
	return nil
}

func (s *TaskCreateSerializer) ToModel() *models.Task {
	const layout = time.RFC3339

	startedAt, _ := time.Parse(layout, s.StartedAt)

	var stoppedAt *time.Time
	if s.StoppedAt != "" {
		parsedStop, err := time.Parse(layout, s.StoppedAt)
		if err == nil {
			stoppedAt = &parsedStop
		}
	}

	return &models.Task{
		BaseSlugUnique: models.BaseSlugUnique{
			Title: s.Title,
		},
		Description: s.Description,
		Mode:        s.Mode,
		Frequency:   s.Frequency,
		StartedAt:   startedAt,
		StoppedAt:   stoppedAt,
		Status:      models.TASK_STATUS_STOPPED,
	}
}

type TaskListResponseSerializer struct {
	BaseSlugResponseSerializer
	Description string `json:"description"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at"`
	StoppedAt   string `json:"stopped_at,omitempty"`
}

// TaskListResponse serialize danh sách task thành slice TaskListResponseSerializer
func TaskListResponse(tasks []models.Task) []TaskListResponseSerializer {
	const layout = "2006-01-02T15:04:05"
	results := make([]TaskListResponseSerializer, len(tasks))

	for i, task := range tasks {
		serializer := TaskListResponseSerializer{
			BaseSlugResponseSerializer: BaseSlugResponseSerializer{
				BaseResponseSerializer: BaseResponseSerializer{
					ID:        task.ID,
					CreatedAt: task.CreatedAt,
					UpdatedAt: task.UpdatedAt,
				},
				Title: task.Title,
				Slug:  task.Slug,
			},
			Description: task.Description,
			Status:      task.Status,
			StartedAt:   task.StartedAt.Format(layout),
		}

		if task.StoppedAt != nil {
			serializer.StoppedAt = task.StoppedAt.Format(layout)
		}

		results[i] = serializer
	}

	return results
}

type TaskDetailResponseSerializer struct {
	BaseSlugResponseSerializer
	Description string `json:"description"`
	Mode        string `json:"mode"`
	Frequency   int    `json:"frequency"`
	StartedAt   string `json:"started_at"`
	StoppedAt   string `json:"stopped_at,omitempty"`
	Status      string `json:"status"`
}

func TaskDetailResponse(instance *models.Task) *TaskDetailResponseSerializer {
	const layout = "2006-01-02T15:04:05"

	var stoppedAt string
	if instance.StoppedAt != nil {
		stoppedAt = instance.StoppedAt.Format(layout)
	}

	return &TaskDetailResponseSerializer{
		BaseSlugResponseSerializer: BaseSlugResponseSerializer{
			BaseResponseSerializer: BaseResponseSerializer{
				ID:        instance.ID,
				CreatedAt: instance.CreatedAt,
				UpdatedAt: instance.UpdatedAt,
			},
			Slug:  instance.Slug,
			Title: instance.Title,
		},
		Description: instance.Description,
		Mode:        instance.Mode,
		Frequency:   instance.Frequency,
		StartedAt:   instance.StartedAt.Format(layout),
		StoppedAt:   stoppedAt,
		Status:      instance.Status,
	}
}
