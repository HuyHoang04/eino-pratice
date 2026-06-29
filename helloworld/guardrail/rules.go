package guardrail

import (
	"fmt"
	"regexp"
	"strings"
)

// injectionRule is one heuristic detector for prompt-injection patterns.
type injectionRule struct {
	desc   string
	pattern *regexp.Regexp
}

// Lowercase, case-insensitive patterns. Both English and Vietnamese triggers.
var injectionRules = []injectionRule{
	{"override/ignore previous instructions",
		regexp.MustCompile(`(?i)(ignore|disregard|forget)\s+(all|previous|prior|above|earlier|the)\s+(instructions?|rules?|directives?|prompts?)`)},
	{"reveal system prompt",
		regexp.MustCompile(`(?i)(reveal|show|print|repeat|output|expose)\s+(your|the)\s+(system\s+)?(prompt|instructions?|initial\s+message|secret)`)},
	{"role override / jailbreak persona",
		regexp.MustCompile(`(?i)(you\s+are\s+now|act\s+as|pretend\s+(you\s+are|to\s+be)|from\s+now\s+on\s+you|enter\s+(developer|dan|jailbreak)\s+mode)`)},
	{"new instruction injection",
		regexp.MustCompile(`(?i)(new\s+instructions?\s*[:\-]|system\s*[:\-]\s*you|override\s+(your\s+)?(rules?|constraints?))`)},
	{"ignore guardrails/policy",
		regexp.MustCompile(`(?i)(bypass|disable|turn\s+off|do\s+not\s+follow)\s+(your\s+)?(guardrails?|safety|polic(?:y|ies)|restrictions?|filters?)`)},
	// Vietnamese
	{"bỏ qua hướng dẫn/lệnh (vi)",
		regexp.MustCompile(`(?i)(bỏ\s+qua|làm\s+ngơ|không\s+quan\s+tâm).{0,20}(hướng\s+dẫn|lệnh|quy\s+tắc|chỉ\s+dẫn|rule)`)},
	{"in/tiết lộ system prompt (vi)",
		regexp.MustCompile(`(?i)(in\s+ra|cho\s+xem|tiết\s+lộ|lặp\s+lại).{0,20}(system\s+prompt|hướng\s+dẫn|hệ\s+thống)`)},
	{"đóng vai / giờ bạn là (vi)",
		regexp.MustCompile(`(?i)(giờ\s+bạn\s+là|đóng\s+vai|từ\s+giờ\s+đóng\s+vai|hãy\s+đóng\s+vai)`)},
}

// DetectInjection returns whether any heuristic matched plus the descriptions.
func DetectInjection(text string) (bool, []string) {
	var hits []string
	for _, r := range injectionRules {
		if r.pattern.MatchString(text) {
			hits = append(hits, r.desc)
		}
	}
	return len(hits) > 0, hits
}

// ---- PII detection & redaction ----

type piiRule struct {
	kind    string
	pattern *regexp.Regexp
	// mask returns the redacted replacement for a given match.
	mask func(string) string
}

var piiRules = []piiRule{
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`),
		func(s string) string {
			at := strings.IndexByte(s, '@')
			if at <= 1 {
				return "***"
			}
			return s[:1] + strings.Repeat("*", at-1) + s[at:]
		}},
	// Vietnamese mobile: 09/08/03/07 + 8 digits, optional +84, spaces/dots/dashes.
	{"phone", regexp.MustCompile(`(?i)(\+?84|0)([\s.\-]?)(3|5|7|8|9)\d([\s.\-]?\d){6,8}`),
		func(s string) string {
			keep := stripPIINonDigit(s)
			if len(keep) <= 4 {
				return "****"
			}
			return keep[:len(keep)-4] + strings.Repeat("*", 4)
		}},
	// Basic credit-card-shaped number (13-16 digits, optional separators).
	{"credit_card", regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
		func(s string) string {
			keep := stripPIINonDigit(s)
			if len(keep) < 4 {
				return "****"
			}
			return strings.Repeat("*", len(keep)-4) + keep[len(keep)-4:]
		}},
}

func stripPIINonDigit(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// DetectPII returns the kinds of PII found in text.
func DetectPII(text string) []string {
	var found []string
	for _, r := range piiRules {
		if r.pattern.MatchString(text) {
			found = append(found, r.kind)
		}
	}
	return found
}

// RedactPII masks every PII match in text and returns the sanitized text plus
// the kinds of PII that were redacted.
func RedactPII(text string) (string, []string) {
	var kinds []string
	out := text
	for _, r := range piiRules {
		if r.pattern.MatchString(out) {
			kinds = append(kinds, r.kind)
			out = r.pattern.ReplaceAllStringFunc(out, r.mask)
		}
	}
	return out, kinds
}

// ValidateInput performs basic input-validation checks (non-circular with the
// classifier): empty / over-long / control characters.
func ValidateInput(text string, maxLen int) error {
	t := strings.TrimSpace(text)
	if t == "" {
		return fmt.Errorf("input is empty")
	}
	if len(t) > maxLen {
		return fmt.Errorf("input too long (%d > %d chars)", len(t), maxLen)
	}
	if strings.ContainsAny(t, "\x00") {
		return fmt.Errorf("input contains null byte")
	}
	return nil
}
