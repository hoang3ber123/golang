package serializers

import (
	"auth-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type EmployeeListResponseSerializer struct {
	ID          uuid.UUID                   `json:"id"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
	Username    string                      `json:"username"`
	Identity    string                      `json:"identity"`
	Name        string                      `json:"name"`
	Email       string                      `json:"email"`
	Dob         time.Time                   `json:"dob"`
	Position    string                      `json:"position"`
	PhoneNumber string                      `json:"phone_number"`
	Contact     string                      `json:"contact"`
	IsActive    bool                        `json:"is_active"`
	Role        *RoleListResponseSerializer `json:"role"`
}

// EmployeeListResponse serialize danh sách Employee thành slice EmployeeListResponseSerializer
func EmployeeListResponse(instance *[]models.Employee) []EmployeeListResponseSerializer {
	results := make([]EmployeeListResponseSerializer, len(*instance)) // Preallocate slice

	for i, val := range *instance {
		// Khởi tạo Role mặc định (rỗng) cho serializer
		var roleResp *RoleListResponseSerializer
		if val.Role != nil {
			role := RoleListResponseSerializer{
				ID:    val.Role.ID,
				Slug:  val.Role.Slug,
				Title: val.Role.Title,
			}
			roleResp = &role // Gán con trỏ nếu Role không nil
		}
		// Copy từng phần tử từ models.Employee vào serializer
		results[i] = EmployeeListResponseSerializer{
			ID:          val.ID,
			CreatedAt:   val.CreatedAt,
			UpdatedAt:   val.UpdatedAt,
			Username:    val.Username,
			Identity:    val.Identity,
			Name:        val.Name,
			Email:       val.Email,
			Dob:         val.Dob,
			Position:    val.Position,
			PhoneNumber: val.PhoneNumber,
			Contact:     val.Contact,
			IsActive:    val.IsActive,
			Role:        roleResp,
		}
	}

	return results
}

type EmployeeDetailResponseSerializer struct {
	BaseResponseSerializer
	Username    string                      `json:"username"`
	Name        string                      `json:"name"`
	Email       string                      `json:"email"`
	Dob         time.Time                   `json:"dob"`
	Position    string                      `json:"position,omitempty"`
	PhoneNumber string                      `json:"phone_number,omitempty"`
	Contact     string                      `json:"contact,omitempty"`
	IsActive    bool                        `json:"is_active"`
	Role        *RoleListResponseSerializer `json:"role"`
}

func EmployeeDetailResponse(employee *models.Employee) *EmployeeDetailResponseSerializer {
	var roleResp RoleListResponseSerializer
	if employee.Role != nil {
		role := &RoleListResponseSerializer{
			Title: employee.Role.Title,
			Slug:  employee.Role.Slug,
			ID:    employee.Role.ID,
		}
		roleResp = *role
	}
	// Khởi tạo Role mặc định (rỗng) cho serializer
	return &EmployeeDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		Username:    employee.Username,
		Name:        employee.Name,
		Email:       employee.Email,
		Dob:         employee.Dob,
		Position:    employee.Position,
		PhoneNumber: employee.PhoneNumber,
		Contact:     employee.Contact,
		IsActive:    employee.IsActive,
		Role:        &roleResp,
	}
}

type EmployeeDecentrializeSerializer struct {
}
