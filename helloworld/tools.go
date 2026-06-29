package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetUserInfoInput is the argument schema for the get_user_info tool.
type GetUserInfoInput struct {
	UserID string `json:"user_id" jsonschema:"description=The ID of the user to look up"`
}

// GetUserInfoOutput is the result of the get_user_info tool.
type GetUserInfoOutput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewGetUserInfoTool builds the get_user_info invokable tool.
// Note: it returns an email, which is intentional — it lets the OutputGuard
// demonstrate sensitive-data masking on the agent's final answer.
func NewGetUserInfoTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_user_info",
		"Look up a user's basic info (name, email) by user ID.",
		func(ctx context.Context, input *GetUserInfoInput) (*GetUserInfoOutput, error) {
			return &GetUserInfoOutput{Name: "Alice", Email: "alice@example.com"}, nil
		},
	)
}

// DeleteAccountInput is the argument schema for the delete_account tool.
type DeleteAccountInput struct {
	UserID string `json:"user_id" jsonschema:"description=The ID of the account to permanently delete"`
}

// NewDeleteAccountTool builds the delete_account tool.
// It is a MOCK (no real deletion) and is HIGH RISK, so it is wrapped with
// human-approval + permission in agent.go.
func NewDeleteAccountTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"delete_account",
		"Permanently delete a user account by ID. This is a HIGH-RISK, destructive action.",
		func(ctx context.Context, input *DeleteAccountInput) (string, error) {
			return fmt.Sprintf("account %s has been deleted (mock, no real data affected)", input.UserID), nil
		},
	)
}
