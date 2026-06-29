package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// generateSuggestions là "Curiosity Generator": đọc đoạn trả lời và sinh đúng 3 câu hỏi
// gợi ý ngắn, hấp dẫn dưới dạng mảng JSON, chỉ trong phạm vi nội dung đoạn văn đó.
func generateSuggestions(ctx context.Context, chatModel model.ToolCallingChatModel, answer string) ([]string, error) {
	if strings.TrimSpace(answer) == "" {
		return nil, nil
	}
	const sys = "Bạn là trợ lý sinh câu hỏi gợi ý. Đọc đoạn văn bản người dùng cung cấp và sinh ra " +
		"ĐÚNG 3 câu hỏi đào sâu, cực kỳ ngắn gọn, hấp dẫn, nhắm tới người trẻ, để kích thích tò mò bấm đọc tiếp. " +
		"Chỉ được hỏi trong phạm vi nội dung đoạn văn đó. " +
		"Định dạng đầu ra BẮT BUỘC là một mảng JSON gồm đúng 3 chuỗi, ví dụ: " +
		`["câu hỏi 1", "câu hỏi 2", "câu hỏi 3"]. ` +
		"Không viết thêm gì ngoài mảng JSON."

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(sys),
		schema.UserMessage(answer),
	})
	if err != nil {
		return nil, err
	}
	return parseSuggestions(resp.Content), nil
}

// parseSuggestions bóc mảng chuỗi JSON từ kết quả mô hình
// (chịu được markdown fence ```json ``` và văn bản thừa quanh mảng).
func parseSuggestions(s string) []string {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end < 0 || end < start {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s[start:end+1]), &arr); err != nil {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, q := range arr {
		if q = strings.TrimSpace(q); q != "" {
			out = append(out, q)
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}
