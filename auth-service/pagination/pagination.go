package pagination

import (
	"auth-service/internal/responses"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Pagination struct for metadata
type Pagination struct {
	Page      int `query:"page" json:"page" default:"1"`
	PageSize  int `query:"page_size" json:"page_size" default:"10"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

// PaginatedResponse wraps data and metadata
type PaginatedResponse struct {
	Pagination *Pagination `json:"pagination"`
}

// PaginateWithGORM handles pagination with optional custom query scope
func PaginateWithGORM[T any](c *fiber.Ctx, query *gorm.DB, modelDest *[]T) (*Pagination, *responses.ErrorResponse) {
	// Parse pagination params
	p := &Pagination{}
	if err := c.QueryParser(p); err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, "Invalid pagination parameters: "+err.Error())
	}

	// Set bounds
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 10
	}

	// Calculate offset
	offset := (p.Page - 1) * p.PageSize

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to count records: "+err.Error())
	}
	if total == 0 {
		return nil, responses.NewErrorResponse(fiber.StatusNotFound, "No objects found")
	}

	// Set pagination metadata
	p.Total = int(total)
	p.TotalPage = (p.Total + p.PageSize - 1) / p.PageSize

	// Check if page exceeds total pages
	if p.Page > p.TotalPage {
		return nil, responses.NewErrorResponse(fiber.StatusBadRequest, fmt.Sprintf("Page %d exceeds total pages %d", p.Page, p.TotalPage))
	}

	// Fetch data
	if err := query.Limit(p.PageSize).Offset(offset).Find(modelDest).Error; err != nil {
		return nil, responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to fetch data: "+err.Error())
	}

	// Return pagination info
	return p, nil
}
