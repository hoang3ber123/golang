package handlers

import (
	"product-service/internal/responses"
	"product-service/internal/services"

	"github.com/gofiber/fiber/v2"
)

func UploadMedia(c *fiber.Ctx) error {
	// get form-data body
	form, err := c.MultipartForm()
	if err != nil {
		return responses.NewSuccessResponse(fiber.StatusBadRequest, "Cannot parse form").Send(c)
	}

	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		return responses.NewSuccessResponse(fiber.StatusBadRequest, "No file found in form").Send(c)
	}

	file := files[0]

	uploadPath, errUpload := services.UploadMedia(file, "uploads")
	if errUpload != nil {
		return errUpload.Send(c)
	}

	return responses.NewSuccessResponse(fiber.StatusOK, uploadPath).Send(c)
}
