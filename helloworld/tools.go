package main

import (
	"context"

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
func NewGetUserInfoTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_user_info",
		"This tool is for checking user information via user ID",
		func(ctx context.Context, input *GetUserInfoInput) (*GetUserInfoOutput, error) {
			return &GetUserInfoOutput{Name: "Alice", Email: "alice@example.com"}, nil
		},
	)
}
