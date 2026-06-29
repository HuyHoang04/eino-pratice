package main

import (
	"context"
	"os"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

// buildChatModel tạo ChatModel (OpenAI-compatible) để tái dùng cho cả Culture Agent
// và Curiosity Generator.
func buildChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		ByAzure: os.Getenv("OPENAI_BY_AZURE") == "true",
	})
}

// buildAgent dựng ChatModelAgent (với cultural_greeter) để tái dùng.
func buildAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (*adk.ChatModelAgent, error) {
	greeter, err := NewCulturalGreeterTool()
	if err != nil {
		return nil, err
	}
	return NewAgent(ctx, chatModel, greeter)
}
