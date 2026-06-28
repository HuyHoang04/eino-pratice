package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
)

// mustRun invokes the tool with a JSON argument and returns the decoded result.
func mustRun[T any](t *testing.T, run func(ctx context.Context, args string, opts ...tool.Option) (string, error), args string) *T {
	t.Helper()
	out, err := run(context.Background(), args)
	if err != nil {
		t.Fatalf("InvokableRun err: %v", err)
	}
	var got T
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", out, err)
	}
	return &got
}

// contains reports whether slice contains s (case-insensitive).
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
