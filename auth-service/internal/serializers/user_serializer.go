package serializers

import (
	"auth-service/internal/models"
	"time"
)

type UserDetailResponseSerializer struct {
	BaseResponseSerializer
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Dob         time.Time `json:"dob"`
	Position    *string   `json:"position,omitempty"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	Contact     *string   `json:"contact,omitempty"`
}

func UserDetailResponse(user *models.User) *UserDetailResponseSerializer {
	return &UserDetailResponseSerializer{
		BaseResponseSerializer: BaseResponseSerializer{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Username:    user.Username,
		Name:        user.Name,
		Email:       user.Email,
		Dob:         user.Dob,
		Position:    user.Position,
		PhoneNumber: user.PhoneNumber,
		Contact:     user.Contact,
	}
}

// UserListResponseSerializer struct để serialize danh sách Role
type UserListResponseSerializer struct {
	BaseResponseSerializer
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// UserListResponseSerializer serialize danh sách Role thành slice UserListResponseSerializer
func UserListResponse(instance *[]models.User) []UserListResponseSerializer {
	results := make([]UserListResponseSerializer, len(*instance)) // Preallocate slice

	for i, val := range *instance {
		// Copy từng phần tử từ models.Role vào serializer
		results[i] = UserListResponseSerializer{
			BaseResponseSerializer: BaseResponseSerializer{
				ID:        val.ID,
				CreatedAt: val.CreatedAt,
				UpdatedAt: val.UpdatedAt,
			},
			Username: val.Username,
			Name:     val.Name,
			Email:    val.Email,
		}
	}

	return results
}
