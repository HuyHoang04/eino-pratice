package main

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ChatResponse là kết quả một lượt chat: câu trả lời + (tối đa) 3 câu hỏi gợi ý.
type ChatResponse struct {
	Answer      string   `json:"answer"`
	Suggestions []string `json:"suggestions"`
}

// buildChatChain dựng một Eino Chain gồm 2 mắt xích và là điểm "Cô lập ngữ cảnh":
//
//  1. Culture Agent: chạy ReAct agent để trả lời; chỉ câu trả lời (string) được truyền
//     sang mắt xích 2 — toàn bộ lịch sử chat / sự kiện trung gian đều bị loại bỏ
//     để không làm nhiễu mô hình sinh gợi ý.
//  2. Curiosity Generator: từ ĐÚNG câu trả lời vừa sinh, sinh 3 câu hỏi gợi ý ngắn,
//     hấp dẫn dưới dạng mảng JSON, chỉ trong phạm vi nội dung câu trả lời.
func buildChatChain(ctx context.Context, agent *adk.ChatModelAgent, chatModel model.ToolCallingChatModel) (compose.Runnable[string, *ChatResponse], error) {
	chain := compose.NewChain[string, *ChatResponse]()

	// Mắt xích 1 — Culture Agent + chốt chặn ngữ cảnh.
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, question string) (*ChatResponse, error) {
		answer, err := agentAnswer(ctx, agent, question)
		if err != nil {
			return nil, err
		}
		return &ChatResponse{Answer: answer}, nil
	}))

	// Mắt xích 2 — Curiosity Generator.
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, r *ChatResponse) (*ChatResponse, error) {
		suggestions, err := generateSuggestions(ctx, chatModel, r.Answer)
		if err != nil {
			return nil, err
		}
		r.Suggestions = suggestions
		return r, nil
	}))

	return chain.Compile(ctx)
}

// agentAnswer chạy ReAct agent với câu hỏi và trả về nội dung tin nhắn Assistant cuối cùng.
// Đây chính là điểm cô lập ngữ cảnh: mọi sự kiện trung gian bị bỏ lại, chỉ text cuối được giữ.
func agentAnswer(ctx context.Context, agent *adk.ChatModelAgent, question string) (string, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: true})
	events := runner.Run(ctx, []adk.Message{schema.UserMessage(question)})

	answer := ""
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg, gerr := event.Output.MessageOutput.GetMessage(); gerr == nil && msg != nil {
				if msg.Role == schema.Assistant && strings.TrimSpace(msg.Content) != "" {
					answer = msg.Content
				}
			}
		}
	}
	return answer, nil
}
