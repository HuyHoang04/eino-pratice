package guardrail

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// PermissionChecker decides whether a tool invocation is allowed for the
// current caller. Return (false, reason) to deny; the reason is surfaced to the
// agent so it can react gracefully.
type PermissionChecker func(ctx context.Context, toolName, argumentsInJSON string) (allowed bool, reason string)

// InvokablePermissionTool wraps a tool with an authorization check.
// It follows the same embed-and-override shape as eino's InvokableApprovableTool:
// Info() delegates, InvokableRun() adds the check before delegating on success.
type InvokablePermissionTool struct {
	tool.InvokableTool
	Check PermissionChecker
}

// NewPermissionTool wraps an invokable tool with the given checker.
func NewPermissionTool(t tool.InvokableTool, check PermissionChecker) *InvokablePermissionTool {
	return &InvokablePermissionTool{InvokableTool: t, Check: check}
}

func (i *InvokablePermissionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return i.InvokableTool.Info(ctx)
}

// InvokableRun enforces the permission policy before running the inner tool.
// A denied call returns a refusal string (not an error), so the agent can fold
// the refusal into its reasoning instead of crashing the loop.
func (i *InvokablePermissionTool) InvokableRun(ctx context.Context, argumentsInJSON string,
	opts ...tool.Option,
) (string, error) {
	info, err := i.Info(ctx)
	if err != nil {
		return "", err
	}
	if i.Check != nil {
		if ok, reason := i.Check(ctx, info.Name, argumentsInJSON); !ok {
			return fmt.Sprintf("⛔ tool '%s' is not permitted: %s", info.Name, reason), nil
		}
	}
	return i.InvokableTool.InvokableRun(ctx, argumentsInJSON, opts...)
}

// AllowAll is a permissive checker (everything allowed). Useful for default role.
func AllowAll() PermissionChecker {
	return func(ctx context.Context, _, _ string) (bool, string) { return true, "" }
}

// AllowSet returns a checker that permits only the given tool names.
func AllowSet(names ...string) PermissionChecker {
	allowed := make(map[string]struct{}, len(names))
	for _, n := range names {
		allowed[n] = struct{}{}
	}
	return func(_ context.Context, toolName, _ string) (bool, string) {
		if _, ok := allowed[toolName]; ok {
			return true, ""
		}
		return false, fmt.Sprintf("tool not in allowlist for current role")
	}
}
