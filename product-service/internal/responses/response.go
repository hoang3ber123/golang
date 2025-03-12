package responses

import (
	"github.com/gofiber/fiber/v2"
)

var (
	ErrUnauthorize = &ErrorResponse{
		Message:    "Unauthorized",
		StatusCode: 401,
	}
	ErrForbiden = &ErrorResponse{
		Message:    "Forbiden",
		StatusCode: 403,
	}
	ErrBlockedOrUnactived = &ErrorResponse{
		Message:    "You account have been blocked or unactived",
		StatusCode: 403,
	}
	ErrInternalSystem = &ErrorResponse{
		Message:    "",
		StatusCode: 500,
	}
	ErrInternalDatabase = &ErrorResponse{
		Message:    "",
		StatusCode: 500,
	}
)

// Cấu trúc chung cho phản hồi lỗi
type ErrorResponse struct {
	Status     string `json:"status"`  // always "error"
	Message    string `json:"message"` // error message
	StatusCode int    `json:"-"`
}

// Hàm gửi lỗi theo chuẩn Fiber
func (errResp *ErrorResponse) Send(c *fiber.Ctx) error {
	return c.Status(errResp.StatusCode).JSON(&ErrorResponse{
		Status:  "error",
		Message: errResp.Message,
	})
}

// NewErrorResponse tạo một instance mới của ErrorResponse
func NewErrorResponse(statusCode int, message string) *ErrorResponse {
	return &ErrorResponse{
		Status:     "error",
		Message:    message,
		StatusCode: statusCode,
	}
}

// SuccessResponse là cấu trúc chung cho phản hồi thành công
type SuccessResponse struct {
	Status     string      `json:"status"` // luôn là "success"
	Data       interface{} `json:"data"`   // dữ liệu trả về
	StatusCode int         `json:"-"`      // mã trạng thái HTTP
}

// Send gửi phản hồi thành công theo chuẩn Fiber
func (s *SuccessResponse) Send(c *fiber.Ctx) error {
	s.Status = "success" // Đảm bảo status luôn là "success"
	return c.Status(s.StatusCode).JSON(s)
}

// NewSuccessResponse tạo một instance mới của SuccessResponse
func NewSuccessResponse(statusCode int, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Status: "success",
		Data:   data,
	}
}
