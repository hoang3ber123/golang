package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/responses"
	internal_utils "product-service/internal/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Valid extend
var allowedExtensions = map[string]string{
	".jpg": "image", ".jpeg": "image", ".png": "image", ".gif": "image", ".webp": "image",
	".bmp": "image", ".svg": "image", // Image
	".mp4": "file_video", ".mov": "file_video", ".avi": "file_video",
	".wmv": "file_video", ".flv": "file_video", ".mkv": "file_video", // Video
	".zip": "download_file", ".rar": "download_file", // Extract file
}

// SanitizeFileName: Remove special char in file name
func sanitizeFileName(fileName *multipart.FileHeader) {
	fileName.Filename = strings.TrimSuffix(fileName.Filename, filepath.Ext(fileName.Filename))

	// loại bỏ ký tự đặc biệt
	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	fileName.Filename = re.ReplaceAllString(fileName.Filename, "_")
}

// Hàm tạo hàng loạt medias
func BulkCreateMedia(files []*multipart.FileHeader, relatedID uuid.UUID, relatedType string, directory string) *responses.ErrorResponse {
	// Create slice medias
	medias := make([]models.Media, len(files))

	// Get XAuthToken
	XAuthToken, err := internal_utils.AuthVstorage()
	if err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Error authorizing vstorage: "+err.Error())
	}

	for media_index, file := range files {
		// Check extend in allowedExtensions
		ext := strings.ToLower(filepath.Ext(file.Filename))
		fileType, ok := allowedExtensions[ext]
		if !ok {
			return responses.NewErrorResponse(fiber.StatusBadRequest, "unsupported file extension: "+ext)
		}

		// Sanitize file name
		sanitizeFileName(file)

		// Adding timestamp to file name for unique file name
		timestamp := time.Now().UnixMilli()
		file.Filename = fmt.Sprintf("%s_%d%s", file.Filename, timestamp, ext)

		// Push file
		dirPath := ""
		if fileType != "download_file" {
			dirPath, _, _ = internal_utils.PushFileToVStorage(XAuthToken, file, directory)
			fmt.Println("image:", dirPath)
		} else {
			dirPath, _, _ = internal_utils.PushFileToDownloadVStorage(XAuthToken, file, directory)
			fmt.Println("download_file:", dirPath)
		}

		// Append new model to medias list
		medias[media_index] = models.Media{
			File:        dirPath,
			FileType:    fileType,
			Status:      models.MediaStatusUsing,
			RelatedID:   relatedID,
			RelatedType: relatedType,
		}
	}
	// Save medias to database
	if err := db.DB.Create(&medias).Error; err != nil {
		return responses.NewErrorResponse(fiber.StatusInternalServerError, "Failed to create medias: "+err.Error())
	}
	return nil // Lưu danh sách medias vào DB
}

func BulkUpdateMedia(status string, mediaIDs []uint) error {
	return db.DB.Model(&models.Media{}).
		Where("id IN ?", mediaIDs).
		Update("status", status).Error
}

// Hàm tạo hàng loạt medias
func UploadMedia(file *multipart.FileHeader, directory string) (string, *responses.ErrorResponse) {

	// Get XAuthToken
	XAuthToken, err := internal_utils.AuthVstorage()
	if err != nil {
		return "", responses.NewErrorResponse(fiber.StatusInternalServerError, "Error authorizing vstorage: "+err.Error())
	}

	// Check extend in allowedExtensions
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileType, ok := allowedExtensions[ext]
	if !ok {
		return "", responses.NewErrorResponse(fiber.StatusBadRequest, "unsupported file extension: "+ext)
	}

	// Sanitize file name
	sanitizeFileName(file)

	// Adding timestamp to file name for unique file name
	timestamp := time.Now().UnixMilli()
	file.Filename = fmt.Sprintf("%s_%d%s", file.Filename, timestamp, ext)

	// Push file
	uploadPath := ""
	if fileType != "download_file" {
		_, uploadPath, _ = internal_utils.PushFileToVStorage(XAuthToken, file, directory)
	} else {
		return "", responses.NewErrorResponse(fiber.StatusBadRequest, "file must be image")

	}
	return uploadPath, nil // Lưu danh sách medias vào DB
}
