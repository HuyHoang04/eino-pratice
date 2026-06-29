package guardrail

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
)

// Config configures the Guard.
type Config struct {
	Model       model.BaseChatModel // classifier model (nil => heuristic-only)
	MaxInputLen int                 // input length cap (0 => 4000)
}

// Guard is the top-level guardrail facade exposing InputGuard / OutputGuard.
type Guard struct {
	classifier *Classifier
	moderator  *Moderator
	maxLen     int
}

// NewGuard builds a Guard from config. The OpenAI Moderator is wired from env.
func NewGuard(cfg Config) *Guard {
	if cfg.MaxInputLen <= 0 {
		cfg.MaxInputLen = 4000
	}
	g := &Guard{
		moderator: NewModerator(),
		maxLen:    cfg.MaxInputLen,
	}
	if cfg.Model != nil {
		g.classifier = NewClassifier(cfg.Model)
	}
	return g
}

// Moderator exposes the underlying moderator (for status reporting in the demo).
func (g *Guard) Moderator() *Moderator { return g.moderator }

// ---- INPUT GUARD: Input -> intent -> risk -> decision ----

// InputGuard runs the full pre-agent assessment pipeline on a raw user input.
func (g *Guard) InputGuard(ctx context.Context, input string) Decision {
	// 1) basic validation
	if err := ValidateInput(input, g.maxLen); err != nil {
		return Decision{
			Allowed:     false,
			Sanitized:   input,
			BlockReason: fmt.Sprintf("input validation failed: %v", err),
			Assessment:  RiskAssessment{Intent: "invalid", RiskLevel: RiskBlock},
		}
	}

	// 2) PII sanitize
	sanitized, piiKinds := RedactPII(input)

	// 3) heuristic injection detection (on the raw input)
	heuristicInj, injReasons := DetectInjection(input)

	// 4) LLM intent classifier (best-effort)
	var cls classifierResult
	if g.classifier != nil {
		cls = g.classifier.Classify(ctx, sanitized)
	} else {
		cls = classifierResult{Intent: "unknown", Topic: "allowed"}
	}

	// 5) moderation
	mod := g.moderator.Check(ctx, sanitized)

	// 6) assemble assessment
	a := RiskAssessment{
		Intent:            cls.Intent,
		Topic:             cls.Topic,
		InjectionDetected: heuristicInj || cls.InjectionDetected,
		InjectionReasons:  injReasons,
		ClassifierSays:    cls.Explanation,
		ModerationFlagged: mod.Flagged,
		FlaggedCategories: mod.Categories,
		PIIDetected:       piiKinds,
	}

	// 7) decide
	g.decideInput(&a, mod.Checked)
	return Decision{
		Allowed:     a.RiskLevel != RiskBlock,
		Sanitized:   sanitized,
		Assessment:  a,
		BlockReason: blockReason(a),
	}
}

// decideInput sets RiskLevel + the decision trace Reasons.
func (g *Guard) decideInput(a *RiskAssessment, modChecked bool) {
	// BLOCK triggers first.
	if a.InjectionDetected {
		a.RiskLevel = RiskBlock
		if len(a.InjectionReasons) > 0 {
			a.AddReason("prompt-injection heuristic matched: %s", strings.Join(a.InjectionReasons, ", "))
		}
		if a.ClassifierSays != "" && g.classifier != nil {
			a.AddReason("classifier: %s", a.ClassifierSays)
		}
		return
	}
	if a.Topic == "forbidden" {
		a.RiskLevel = RiskBlock
		if a.ClassifierSays != "" {
			a.AddReason("classifier flagged forbidden topic: %s", a.ClassifierSays)
		} else {
			a.AddReason("classifier flagged forbidden topic")
		}
		return
	}
	if a.ModerationFlagged {
		a.RiskLevel = RiskBlock
		a.AddReason("moderation flagged categories: %s", strings.Join(a.FlaggedCategories, ", "))
		return
	}

	// ALLOWED: assign risk by intent sensitivity.
	a.RiskLevel = intentRisk(a.Intent)
	if len(a.PIIDetected) > 0 {
		a.AddReason("PII redacted before sending to model: %s", strings.Join(a.PIIDetected, ", "))
		// redacting PII is safe, but bump awareness one notch (cap at medium).
		if a.RiskLevel == RiskLow {
			a.RiskLevel = RiskMedium
		}
	}
	if !modChecked {
		a.AddReason("moderation not active (disabled) — relying on heuristic + classifier")
	}
}

// intentRisk maps an intent label to a baseline risk.
func intentRisk(intent string) RiskLevel {
	switch intent {
	case "delete_account", "delete_user", "transfer_money", "refund":
		return RiskHigh
	case "lookup_user", "get_user_info", "smalltalk", "faq":
		return RiskLow
	default:
		return RiskMedium
	}
}

// blockReason composes a user-facing block message from the assessment.
func blockReason(a RiskAssessment) string {
	if a.RiskLevel != RiskBlock {
		return ""
	}
	switch {
	case a.InjectionDetected:
		return "⛔ Blocked: prompt injection detected"
	case a.Topic == "forbidden":
		return "⛔ Blocked: request violates topic/safety policy"
	case a.ModerationFlagged:
		return fmt.Sprintf("⛔ Blocked: moderation flagged [%s]", strings.Join(a.FlaggedCategories, ", "))
	default:
		return "⛔ Blocked: failed guardrail"
	}
}

// ---- OUTPUT GUARD ----

// OutputGuard runs moderation + PII-leak check on the agent's final answer.
func (g *Guard) OutputGuard(ctx context.Context, output string) Decision {
	leaked := DetectPII(output)
	mod := g.moderator.Check(ctx, output)

	a := RiskAssessment{
		Intent:            "agent_response",
		ModerationFlagged: mod.Flagged,
		FlaggedCategories: mod.Categories,
		PIIDetected:       leaked,
	}

	blocked := false
	sanitized := output
	if len(leaked) > 0 {
		blocked = true
		sanitized, _ = RedactPII(output)
		a.AddReason("PII leaked in agent output: %s (masked before return)", strings.Join(leaked, ", "))
	}
	if mod.Flagged {
		blocked = true
		a.AddReason("moderation flagged output categories: %s", strings.Join(mod.Categories, ", "))
	}

	if blocked {
		a.RiskLevel = RiskBlock
		return Decision{
			Allowed:     false,
			Sanitized:   sanitized,
			Assessment:  a,
			BlockReason: "⛔ Output blocked & sanitized (sensitive/policy)",
		}
	}
	a.RiskLevel = RiskLow
	return Decision{Allowed: true, Sanitized: output, Assessment: a}
}

// init silences unused-import worries if log is only used under build tags.
var _ = log.Printf
