package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	einoTool "github.com/cloudwego/eino-examples/adk/common/tool"

	"helloworld/guardrail"
)

// rolePolicy is the tool-permission policy. It reads the caller's role from ctx
// (set via guardrail.WithRole before runner.Query) and gates destructive tools.
// Fail-safe: a missing/unknown role defaults to "guest" (least privilege).
func rolePolicy(ctx context.Context, name, _ string) (bool, string) {
	role := guardrail.RoleFromCtx(ctx)
	switch name {
	case "delete_account":
		if role != "admin" {
			return false, fmt.Sprintf("requires admin role (current role: %s)", role)
		}
	}
	return true, ""
}

// NewAgent builds the Customer-Support ChatModelAgent with guardrail-wrapped tools:
//   - get_user_info -> permission wrapper (role-gated)
//   - delete_account -> approval wrapper (human-in-the-loop) INSIDE permission wrapper
//     so a denied role never even interrupts the human.
func NewAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (*adk.ChatModelAgent, error) {
	userInfo, err := NewGetUserInfoTool()
	if err != nil {
		return nil, fmt.Errorf("new get_user_info tool: %w", err)
	}
	deleteAccount, err := NewDeleteAccountTool()
	if err != nil {
		return nil, fmt.Errorf("new delete_account tool: %w", err)
	}

	// delete_account: human approval first, then role permission on top.
	approvableDelete := &einoTool.InvokableApprovableTool{InvokableTool: deleteAccount}
	wrappedDelete := guardrail.NewPermissionTool(approvableDelete, rolePolicy)

	wrappedUserInfo := guardrail.NewPermissionTool(userInfo, rolePolicy)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "support_agent",
		Description: "A customer-support agent that can look up and delete user accounts.",
		Instruction: `You are a customer-support agent.
You have two tools: "get_user_info" and "delete_account".
- When the user asks to look up a user, call "get_user_info".
- When the user asks to delete/remove an account, call "delete_account".
- Always use the tool for these actions; do not answer from imagination.
- If a tool returns a refusal (starts with "⛔"), report that refusal to the user and do not retry.
Keep replies short.`,
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{wrappedUserInfo, wrappedDelete},
			},
		},
		MaxIterations: 6,
	})
}
