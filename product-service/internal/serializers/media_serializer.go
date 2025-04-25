package serializers

import (
	"fmt"
	"product-service/config"
	"product-service/internal/models"
)

type MediaListResponseSerializer struct {
	ID       uint   `json:"id"`
	FileType string `json:"file_type"`
	File     string `json:"file"`
}

func MediaListResponse(instance *[]models.Media) []MediaListResponseSerializer {
	results := make([]MediaListResponseSerializer, len(*instance)) // Preallocate slice
	baseURL := config.Config.VstorageBaseURL
	baseDownLoadURL := config.Config.VstorageDownloadBaseURL
	for i, val := range *instance {
		url := fmt.Sprintf("%s/%s", baseURL, val.File)

		if val.FileType == "download_file" {
			url = fmt.Sprintf("%s/%s", baseDownLoadURL, val.File)
		}
		// Copy từng phần tử từ models.Category vào serializer
		results[i] = MediaListResponseSerializer{
			ID:       val.ID,
			FileType: val.FileType,
			File:     url,
		}
	}

	return results
}
