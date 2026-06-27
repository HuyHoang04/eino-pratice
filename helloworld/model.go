package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"
)

// GetChatModel initializes a ChatModel based on the MODEL_TYPE environment variable.
// This mimics the factory pattern used in eino-examples.
func GetChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	modelType := os.Getenv("MODEL_TYPE")
	if modelType == "" {
		modelType = "gemini"
	}

	switch modelType {
	case "openai":
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Model:   os.Getenv("OPENAI_MODEL"),
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			ByAzure: func() bool {
				return os.Getenv("OPENAI_BY_AZURE") == "true"
			}(),
		})
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		baseURL := os.Getenv("GEMINI_BASE_URL")
		clientConfig := &genai.ClientConfig{
			APIKey: apiKey,
		}
		if baseURL != "" {
			clientConfig.HTTPOptions = genai.HTTPOptions{
				BaseURL: baseURL,
			}
		}
		client, err := genai.NewClient(ctx, clientConfig)
		if err != nil {
			log.Fatalf("NewClient of gemini failed, err=%v", err)
			return nil, err
		}

		modelName := os.Getenv("GEMINI_MODEL")
		if modelName == "" {
			modelName = "gemini-2.5-flash-lite"
		}

		return gemini.NewChatModel(ctx, &gemini.Config{
			Client: client,
			Model:  modelName,
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: true,
				ThinkingBudget:  nil,
			},
		})
	default:
		return nil, fmt.Errorf("unsupported MODEL_TYPE: %s", modelType)
	}
}
