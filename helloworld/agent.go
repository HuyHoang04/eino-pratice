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
		Name:        "hello_agent",
		Description: "A friendly greeting assistant",
		Instruction: "You are a friendly assistant. Please respond to the user in a warm tone.",
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 10,
	})
}
