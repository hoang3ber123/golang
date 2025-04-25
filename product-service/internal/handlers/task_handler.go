package handlers

import (
	"context"
	"errors"
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	"product-service/internal/serializers"
	"product-service/internal/tasks"
	openai "product-service/open-ai"
	"product-service/pagination"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TaskCreate(c *fiber.Ctx) error {
	// Xử lý tạo category nếu xác thực thành công
	serializer := new(serializers.TaskCreateSerializer)
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}
	// Serializer to model
	task := serializer.ToModel()
	fmt.Println(task)
	if err := db.DB.Create(&task).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to create task: "+err.Error()).Send(c)
	}
	// Response
	return responses.NewSuccessResponse(fiber.StatusCreated, "Create task successfully").Send(c)
}

func TaskList(c *fiber.Ctx) error {
	titleQuery := c.Query("title")
	createdAtOrderQuery := c.Query("created_at_order") // "asc" hoặc "desc"

	query := db.DB.Model(&models.Task{})

	// Lọc theo tiêu đề (title)
	if titleQuery != "" {
		query = query.Where("title LIKE ?", "%"+titleQuery+"%")
	}

	// Sắp xếp theo created_at nếu có
	if createdAtOrderQuery != "" {
		order := strings.ToLower(createdAtOrderQuery)
		if order == "asc" || order == "desc" {
			query = query.Order("created_at " + order)
		} else {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid created_at_order, use 'asc' or 'desc'").Send(c)
		}
	}

	// Phân trang
	var tasks []models.Task
	paginator, err := pagination.PaginateWithGORM(c, query, &tasks)
	if err != nil {
		return err.Send(c)
	}

	var result interface{}
	if tasks != nil {
		result = serializers.TaskListResponse(tasks)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     result,
	}).Send(c)
}

func TaskDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var instance models.Task
	if err := db.DB.First(&instance, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return responses.NewErrorResponse(fiber.StatusNotFound, "task not found").Send(c)
		}
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Database error: "+err.Error()).Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, serializers.TaskDetailResponse(&instance)).Send(c)
}

func RunTask(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid task ID").Send(c)
	}

	var task models.Task
	if err := db.DB.First(&task, "id = ?", id).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusNotFound, "Task not found").Send(c)
	}
	fmt.Println(task)
	if task.Status == models.TASK_STATUS_RUNNING {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Task is already running").Send(c)
	}

	if task.StoppedAt != nil && time.Now().After(*task.StoppedAt) {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Task execution time has expired").Send(c)
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		task.Status = models.TASK_STATUS_RUNNING
		return tx.Save(&task).Error
	})
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to update task").Send(c)
	}

	prompt := tasks.GeneratePromptForCreateProduct(task.Description)

	tasks.RunTask(&task, func(ctx context.Context) error {
		resp, err := openai.AskGemini(prompt)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return err
		}
		fmt.Println(resp)
		return nil
	})

	return responses.NewSuccessResponse(fiber.StatusOK, "Task started").Send(c)
}

func StopTask(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid task ID").Send(c)
	}

	stopped := tasks.StopTask(id)
	if !stopped {
		return responses.NewErrorResponse(fiber.StatusNotFound, "Task is not running or not found").Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, "Task stopped").Send(c)
}
