package tasks

import (
	"encoding/json"
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"product-service/internal/services"
	"strings"

	"github.com/google/uuid"
)

func UpdateProductTaskFromAI() {

}

func GeneratePromptForCreateProduct(modifyQuery string) string {
	// Fetch categories from Redis
	categoriesJSON, err := services.GetCategoriesFromRedis()
	if err != nil {
		fmt.Println("failed to fetch categories:", err)
		return ""
	}

	// Parse categories
	var allCategories []models.Category
	if err := json.Unmarshal([]byte(categoriesJSON), &allCategories); err != nil {
		fmt.Println("failed to unmarshal categories:", err.Error())
		return ""
	}

	// Build category list for prompt
	var categoryList strings.Builder
	categoryList.WriteString("Available categories:\n")
	for _, cat := range allCategories {
		if cat.ID == uuid.Nil {
			continue
		}
		categoryList.WriteString(fmt.Sprintf("- ID: %s, Title: %s\n", cat.ID.String(), cat.Title))
	}

	// Construct prompt
	prompt := fmt.Sprintf(`
Generate a list of products based on the following query: "%s".
Return the response as a JSON array of objects, where each object has:
- "title": a string (non-empty, max 255 characters)
- "description": a string (non-empty, max 1000 characters)
- "category_ids": an array of category IDs (strings, must match IDs from the provided category list)

%s

Example output:
[
    {
        "title": "Smartphone",
        "description": "A high-end smartphone with advanced features.",
        "category_ids": ["%s"]
    }, ...
]

Ensure:
- The JSON is valid and well-formed.
- Category IDs are valid UUIDs from the provided list.
- Titles and descriptions are relevant to the query.
`, modifyQuery, categoryList.String(), allCategories[0].ID.String())

	return prompt
}

func CreateProductTaskFromAI(response string, userID uuid.UUID) error {
	type AIProductResponse struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		CategoryIDs []string `json:"category_ids"`
	}

	var aiProducts []AIProductResponse
	if err := json.Unmarshal([]byte(response), &aiProducts); err != nil {
		fmt.Println("failed to parse JSON:", err)
	}

	for _, aiProduct := range aiProducts {
		// Validate required fields
		if strings.TrimSpace(aiProduct.Title) == "" {
			continue // Skip products with empty titles
		}

		// Fetch categories
		var categories []models.Category
		for _, catID := range aiProduct.CategoryIDs {
			parsedID, err := uuid.Parse(catID)
			if err != nil {
				// Log invalid category ID and skip
				fmt.Printf("Invalid category ID %s: %v\n", catID, err)
				continue
			}
			var category models.Category
			if err := db.DB.Where("id = ?", parsedID).First(&category).Error; err != nil {
				// Log missing category and skip
				fmt.Printf("Category ID %s not found: %v\n", catID, err)
				continue
			}
			categories = append(categories, category)
		}

		// Create product
		product := models.Product{
			BaseSlug: models.BaseSlug{
				Title: aiProduct.Title,
			},
			Description: aiProduct.Description,
			UserID:      userID,
			Link:        nil,
			Price:       nil,
			Categories:  categories,
		}

		if err := db.DB.Create(&product).Error; err != nil {
			fmt.Println("failed to create product:", err.Error())
		}
	}

	return nil
}
