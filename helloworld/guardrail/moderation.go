package guardrail

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	modopenai "github.com/meguminnnnnnnnn/go-openai"
)

// ModerationResult is the normalized output of a moderation check.
type ModerationResult struct {
	Checked   bool     // false when moderation is disabled (no API key)
	Flagged   bool
	Categories []string
}

// Moderator wraps the OpenAI moderation endpoint.
// It is optional: if no API key is configured, Check degrades to "not checked".
type Moderator struct {
	client  *modopenai.Client
	model   string
	enabled bool
}

// NewModerator builds a Moderator from env. Requires OPENAI_API_KEY to be enabled.
// OPENAI_MODERATION_MODEL is optional (defaults to omni-moderation-latest).
// OPENAI_MODERATION_BASE_URL is optional (for proxies / Azure).
func NewModerator() *Moderator {
	m := &Moderator{model: modopenai.ModerationOmniLatest}
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		log.Printf("[guardrail] moderation DISABLED (OPENAI_API_KEY not set); will skip moderation checks")
		return m
	}
	if mm := strings.TrimSpace(os.Getenv("OPENAI_MODERATION_MODEL")); mm != "" {
		m.model = mm
	}
	cfg := modopenai.DefaultConfig(key)
	if base := strings.TrimSpace(os.Getenv("OPENAI_MODERATION_BASE_URL")); base != "" {
		cfg.BaseURL = base
	}
	m.client = modopenai.NewClientWithConfig(cfg)
	m.enabled = true
	log.Printf("[guardrail] moderation ENABLED (model=%s)", m.model)
	return m
}

// Enabled reports whether moderation will actually call the API.
func (m *Moderator) Enabled() bool { return m.enabled }

// Check calls the OpenAI moderation API on text.
func (m *Moderator) Check(ctx context.Context, text string) ModerationResult {
	if !m.enabled {
		return ModerationResult{Checked: false}
	}
	resp, err := m.client.Moderations(ctx, modopenai.ModerationRequest{
		Input: text,
		Model: m.model,
	})
	if err != nil {
		// Degrade gracefully: a failed moderation call must not break the agent.
		log.Printf("[guardrail] moderation call failed: %v", err)
		return ModerationResult{Checked: false}
	}
	if len(resp.Results) == 0 {
		return ModerationResult{Checked: true}
	}
	r := resp.Results[0]
	var cats []string
	if r.Flagged {
		cats = flaggedCategories(r.Categories)
	}
	return ModerationResult{Checked: true, Flagged: r.Flagged, Categories: cats}
}

// flaggedCategories returns the names of categories that were triggered.
func flaggedCategories(c modopenai.ResultCategories) []string {
	out := make([]string, 0, 11)
	pairs := []struct {
		name string
		val  bool
	}{
		{"hate", c.Hate}, {"hate/threatening", c.HateThreatening},
		{"harassment", c.Harassment}, {"harassment/threatening", c.HarassmentThreatening},
		{"self-harm", c.SelfHarm}, {"self-harm/intent", c.SelfHarmIntent},
		{"self-harm/instructions", c.SelfHarmInstructions},
		{"sexual", c.Sexual}, {"sexual/minors", c.SexualMinors},
		{"violence", c.Violence}, {"violence/graphic", c.ViolenceGraphic},
	}
	for _, p := range pairs {
		if p.val {
			out = append(out, p.name)
		}
	}
	return out
}

// Reason is a small helper for composing block reasons.
func reasonf(format string, args ...any) string { return fmt.Sprintf(format, args...) }
