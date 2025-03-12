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
	for i, val := range *instance {
		// Copy từng phần tử từ models.Category vào serializer
		results[i] = MediaListResponseSerializer{
			ID:       val.ID,
			FileType: val.FileType,
			File:     fmt.Sprintf("%s/%s", baseURL, val.File),
		}
	}

	return results
}
