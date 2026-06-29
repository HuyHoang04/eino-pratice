package guardrail

import "context"

// roleKey is the context key for the caller's role, used by permission policies.
type roleKey struct{}

// WithRole returns a ctx carrying the caller's role. The runner propagates this
// ctx down to tool.InvokableRun, so the permission wrapper can read it.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey{}, role)
}

// RoleFromCtx returns the caller's role, defaulting to "guest" (least privilege)
// when none is set. Fail-safe: unknown caller never gets privileged access.
func RoleFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(roleKey{}).(string); ok && v != "" {
		return v
	}
	return "guest"
}
