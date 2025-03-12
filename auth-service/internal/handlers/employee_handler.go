package handlers

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/responses"
	"auth-service/internal/serializers"
	"auth-service/pagination"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func EmployeeCreate(c *fiber.Ctx) error {
	var serializer serializers.EmployeeListSignUpSerializer

	// Kiểm tra dữ liệu đầu vào
	if err := serializer.IsValid(c); err != nil {
		return err.Send(c)
	}

	// Chuyển sang model
	employees, err := serializer.ToModel()
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Thực hiện các thao tác cơ sở dữ liệu trong giao dịch (dùng 'tx' thay vì 'db')
		// Lưu vào database (giả sử dùng GORM)
		if err := tx.Create(&employees).Error; err != nil {
			return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
		}
		// Trả về nil sẽ commit (xác nhận) toàn bộ giao dịch
		return nil
	}); err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, err.Error()).Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, "Create Successfully").Send(c)
}

func EmployeeList(c *fiber.Ctx) error {
	// Khởi tạo truy vấn cơ bản
	query := db.DB.Model(&models.Employee{}).Joins("Role")
	// Khai báo struct query
	params := new(models.EmployeeQuery)
	// Lấy các tham số truy vấn từ query string
	if err := c.QueryParser(params); err != nil {
		return responses.NewErrorResponse(fiber.StatusBadRequest, err.Error()).Send(c)
	}

	// Áp dụng điều kiện lọc nếu tham số không rỗng
	if params.Identity != "" {
		query = query.Where("identity = ?", params.Identity)
	}
	if params.Username != "" {
		query = query.Where("username = ?", params.Username)
	}
	if params.Email != "" {
		query = query.Where("email LIKE ?", "%"+params.Email+"%")
	}
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%") // Tìm kiếm gần đúng với name
	}
	if params.Position != "" {
		query = query.Where("position LIKE ?", "%"+params.Position+"%")
	}
	if params.PhoneNumber != "" {
		query = query.Where("phone_number = ?", params.PhoneNumber)
	}
	if c.Query("is_active") != "" { // Kiểm tra nếu tham số is_active được truyền
		query = query.Where("is_active = ?", params.IsActive)
	}
	// Lọc theo title của Role thay vì role_id
	if params.RoleTitle != "" {
		query = query.Where("Role.title LIKE ?", "%"+params.RoleTitle+"%")
	}

	// Sử dụng hàm phân trang
	var employees []models.Employee
	paginator, err := pagination.PaginateWithGORM(c, query, &employees)
	if err != nil {
		return err.Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, fiber.Map{
		"pagination": paginator,
		"result":     serializers.EmployeeListResponse(&employees),
	}).Send(c)
}

func EmployeeDetail(c *fiber.Ctx) error {
	employee := c.Locals("employee").(*models.Employee)
	return responses.NewSuccessResponse(fiber.StatusOK, serializers.EmployeeDetailResponse(employee)).Send(c)
}
