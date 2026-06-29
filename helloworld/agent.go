package main

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewAgent builds the ReAct ChatModelAgent.
// ChatModelAgent runs the ReAct loop when ToolsConfig is set (Reason -> Action -> Act -> Observe).
func NewAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools ...tool.BaseTool) (*adk.ChatModelAgent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "culture_agent",
		Description: "Chatbot Agent bảo tồn văn hóa Việt Nam cho người trẻ — tiếp đón & giảng giải kiến thức",
		Instruction: `Bạn là một Chatbot Agent bảo tồn văn hóa Việt Nam, dành cho đối tượng người trẻ.
Bạn đảm nhận hai vai trò: tiếp đón (onboarding) và giảng giải kiến thức văn hóa.

Quy tắc:
1. Khi người dùng chào hỏi hoặc tự giới thiệu tên (ví dụ: "Chào bot, mình là Linh."),
   BẮT BUỘC gọi công cụ cultural_greeter và truyền đúng tên của họ vào trường name,
   sau đó dùng kết quả công cụ trả về làm lời chào. Có thể thêm 1-2 câu mời gần gũi với giới trẻ.
2. Khi người dùng hỏi về kiến thức văn hóa (lịch sử, ý nghĩa, phong tục, trang phục, lễ hội...),
   hãy trả lời CHI TIẾT, chính xác, đúng trọng tâm bằng tiếng Việt — tập trung đúng chủ đề câu hỏi.
3. Tuyệt đối không bịa đặt kiến thức văn hóa.
4. Luôn trả lời bằng tiếng Việt, giọng điệu trẻ trung, ấm áp.`,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
	})
}
