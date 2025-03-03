package responses

import "github.com/gofiber/fiber/v2"

// Cấu trúc chung cho phản hồi lỗi
type ErrorResponse struct {
	Status  string `json:"status"`  // always "error"
	Message string `json:"message"` // error message
}

// Cấu trúc chung cho phản hồi thành công
type SuccessResponse struct {
	Status string      `json:"status"` // always "success"
	Data   interface{} `json:"data"`   // dữ liệu trả về (có thể là object, list, map...)
}

// Hàm gửi lỗi theo chuẩn Fiber
func SendErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

// Hàm gửi response thành công theo chuẩn Fiber
func SendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(&SuccessResponse{
		Status: "success",
		Data:   data,
	})
}
