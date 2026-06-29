package guardrail

import (
	"fmt"
	"strings"
)

// RiskLevel is the severity assigned by the risk assessor.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
	RiskBlock  RiskLevel = "blocked"
)

// RiskAssessment is the result of "analyzing the situation": what the user most
// likely wants, whether it is on-topic, and which guardrail signals fired.
// It is the heart of the Input -> intent -> risk -> decision flow.
type RiskAssessment struct {
	Intent            string   // lookup_user | delete_account | smalltalk | off_topic | jailbreak | ...
	Topic             string   // "allowed" | "forbidden"
	InjectionDetected bool
	InjectionReasons  []string // why we think it's an injection (heuristic matches)
	ClassifierSays    string   // free-text explanation from the LLM classifier
	ModerationFlagged bool
	FlaggedCategories []string // hate, violence, ...
	PIIDetected       []string // email, phone, credit_card
	RiskLevel         RiskLevel
	Reasons           []string // human-readable trace, used to explain the demo
}

// AddReason appends a reason line to the assessment's decision trace.
func (a *RiskAssessment) AddReason(format string, args ...any) {
	a.Reasons = append(a.Reasons, fmt.Sprintf(format, args...))
}

// String renders a compact, demo-friendly block of the assessment.
func (a RiskAssessment) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  intent           : %s\n", a.Intent)
	fmt.Fprintf(&b, "  topic            : %s\n", a.Topic)
	fmt.Fprintf(&b, "  injection        : %v", a.InjectionDetected)
	if len(a.InjectionReasons) > 0 {
		fmt.Fprintf(&b, "  (%s)", strings.Join(a.InjectionReasons, ", "))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "  classifier note  : %s\n", orDash(a.ClassifierSays))
	fmt.Fprintf(&b, "  moderation       : %v", a.ModerationFlagged)
	if len(a.FlaggedCategories) > 0 {
		fmt.Fprintf(&b, "  [%s]", strings.Join(a.FlaggedCategories, ", "))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "  pii              : %v", len(a.PIIDetected) > 0)
	if len(a.PIIDetected) > 0 {
		fmt.Fprintf(&b, "  [%s]", strings.Join(a.PIIDetected, ", "))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "  risk             : %s\n", a.RiskLevel)
	if len(a.Reasons) > 0 {
		b.WriteString("  decision trace   :\n")
		for _, r := range a.Reasons {
			fmt.Fprintf(&b, "    - %s\n", r)
		}
	}
	return b.String()
}

// Decision is what InputGuard / OutputGuard hand back to the driver.
type Decision struct {
	Allowed     bool
	Sanitized   string // input/output after PII redaction
	Assessment  RiskAssessment
	BlockReason string
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
