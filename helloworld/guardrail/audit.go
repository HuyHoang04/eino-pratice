package guardrail

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AuditCallback is an observability handler: it logs every message going INTO
// the chat model and every message coming OUT. It deliberately does NOT block
// or mutate anything (eino callbacks cannot block a run by design).
//
// This is the "detection" layer — distinct from the enforcement layers
// (InputGuard / OutputGuard / tool wrappers). It proves we understand where
// eino lets us observe vs. where it lets us act.
type AuditCallback struct {
	logFn func(format string, args ...any)
}

// NewAuditCallback builds an audit handler. logFn defaults to log.Printf.
func NewAuditCallback() *AuditCallback {
	return &AuditCallback{logFn: log.Printf}
}

// Needed opts into OnStart / OnEnd / OnError / OnEndWithStreamOutput only,
// so the framework skips stream-input copying overhead.
func (a *AuditCallback) Needed(_ context.Context, _ *callbacks.RunInfo, timing callbacks.CallbackTiming) bool {
	switch timing {
	case callbacks.TimingOnStart, callbacks.TimingOnEnd,
		callbacks.TimingOnError, callbacks.TimingOnEndWithStreamOutput:
		return true
	default:
		return false
	}
}

func (a *AuditCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, in callbacks.CallbackInput) context.Context {
	if !relevantNode(info.Name) {
		return ctx
	}
	if mi := model.ConvCallbackInput(in); mi != nil {
		if t := lastUserText(mi.Messages); t != "-" {
			a.logFn("[AUDIT -> %s] %s", nodeLabel(info.Name), t)
		}
	}
	return ctx
}

func (a *AuditCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, out callbacks.CallbackOutput) context.Context {
	if !relevantNode(info.Name) {
		return ctx
	}
	if mo := model.ConvCallbackOutput(out); mo != nil && mo.Message != nil {
		a.logFn("[AUDIT <- %s] %s", nodeLabel(info.Name), preview(mo.Message.Content))
	}
	return ctx
}

func (a *AuditCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	// Interrupts surface as errors in eino; they are expected flow control here,
	// not real failures, so we suppress the noisy internal signal in the audit log.
	if isInterruptErr(err) {
		return ctx
	}
	a.logFn("[AUDIT !! %s] error: %v", nodeLabel(info.Name), err)
	return ctx
}

func (a *AuditCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	_ *schema.StreamReader[callbacks.CallbackInput],
) context.Context {
	return ctx
}

// OnEndWithStreamOutput logs streamed model output by consuming the copied
// stream the framework handed us. The copy MUST be closed.
func (a *AuditCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	out *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {
	defer out.Close()
	if !relevantNode(info.Name) {
		return ctx
	}
	var b strings.Builder
	for {
		chunk, err := out.Recv()
		if err != nil {
			break
		}
		if mo := model.ConvCallbackOutput(chunk); mo != nil && mo.Message != nil {
			b.WriteString(mo.Message.Content)
		}
	}
	if b.Len() > 0 {
		a.logFn("[AUDIT <- %s] %s", nodeLabel(info.Name), preview(b.String()))
	}
	return ctx
}

// lastUserText returns the content of the most recent user message, if any.
func lastUserText(msgs []*schema.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i] != nil && msgs[i].Role == schema.User {
			return preview(msgs[i].Content)
		}
	}
	if len(msgs) > 0 && msgs[len(msgs)-1] != nil {
		return preview(msgs[len(msgs)-1].Content)
	}
	return "-"
}

func preview(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	const max = 200
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// RegisterAudit installs an audit callback globally. Call once at startup.
func RegisterAudit() {
	callbacks.AppendGlobalHandlers(NewAuditCallback())
}

// nodeLabel prettifies the internal node name for the demo log.
func nodeLabel(name string) string {
	switch name {
	case "", "ChatModel":
		return "classifier/agent"
	case "ReAct":
		return "react-loop"
	default:
		return name
	}
}

// relevantNode suppresses eino's internal plumbing nodes (cancel-checks,
// tool-node machinery) that only duplicate or clutter the audit trail.
func relevantNode(name string) bool {
	switch name {
	case "CancelCheck", "AfterToolCallsCancelCheck", "ToolNode":
		return false
	default:
		return true
	}
}

// isInterruptErr detects eino interrupt signals surfaced via OnError.
func isInterruptErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "interrupt")
}
