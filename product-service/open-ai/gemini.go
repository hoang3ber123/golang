package openai

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	GeminiModel *genai.GenerativeModel
	once        sync.Once
)

// InitGeminiModel initializes the Gemini GenerativeModel singleton.
func InitGeminiModel() {
	once.Do(func() {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
		if err != nil {
			log.Fatalf("❌ Failed to initialize Gemini client: %v", err)
		}

		model := client.GenerativeModel("gemini-1.5-flash")
		GeminiModel = model

		log.Println("✅ Gemini model initialized successfully")
	})
}

// Chuyển kết quả trả về thành string
func GeminiResponseToString(resp *genai.GenerateContentResponse) string {
	respString := ""
	for _, candidate := range resp.Candidates {
		if candidate != nil {
			if candidate.Content.Parts != nil {
				respString = respString + "\n" + string(candidate.Content.Parts[0].(genai.Text))
			}
		}
	}

	return respString
}

func AskGemini(prompt string) (string, error) {
	resp, err := GeminiModel.GenerateContent(context.Background(), genai.Text(prompt))
	if err != nil {
		fmt.Printf("❌ Error generating content: %v", err)
		return "", err
	}
	return GeminiResponseToString(resp), nil
}
