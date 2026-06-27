package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-examples/adk/common/prints"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	ctx := context.Background()

	model, err := GetChatModel(ctx)
	if err != nil {
		log.Fatalf("GetChatModel failed: %v", err)
	}

	getUserInfoTool, err := NewGetUserInfoTool()
	if err != nil {
		log.Fatalf("NewGetUserInfoTool failed: %v", err)
	}

	agent, err := NewAgent(ctx, model, getUserInfoTool)
	if err != nil {
		log.Fatalf("NewAgent failed: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	input := []adk.Message{
		schema.UserMessage("Please look up user information for user id 12345."),
	}

	events := runner.Run(ctx, input)

	for {
		event, ok := events.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("Agent error: %v", event.Err)
			break
		}

		prints.Event(event)
	}
}
