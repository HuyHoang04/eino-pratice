package main

import (
	"strings"
	"testing"
	"time"
)

func TestComposeGreeting_FormatWithGivenName(t *testing.T) {
	got := composeGreeting("Linh", "vào thời Nguyễn người xưa hay chắp tay")
	if !strings.HasPrefix(got, "Chào Linh.") {
		t.Fatalf("expected greeting to start with 'Chào Linh.', got: %q", got)
	}
	if !strings.Contains(got, "Bạn có biết") || !strings.HasSuffix(got, "Hôm nay bạn muốn tìm hiểu về nét văn hóa nào?") {
		t.Fatalf("expected 'Bạn có biết ...? Hôm nay bạn muốn tìm hiểu ...?' frame, got: %q", got)
	}
}

func TestComposeGreeting_FallsBackWhenNameEmpty(t *testing.T) {
	got := composeGreeting("", "một sự thật văn hóa")
	if !strings.HasPrefix(got, "Chào bạn.") {
		t.Fatalf("expected fallback 'Chào bạn.', got: %q", got)
	}
}

func TestPickCulturalContent_DeterministicPerDayAndFromPool(t *testing.T) {
	now := time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC)
	first := pickCulturalContent(now)
	second := pickCulturalContent(now)
	if first != second {
		t.Fatalf("expected deterministic content for the same day, got %q then %q", first, second)
	}
	if strings.TrimSpace(first) == "" {
		t.Fatal("expected non-empty cultural content")
	}
	pool := append(append([]string{}, greetingCustoms...), culturalFacts...)
	if !containsString(pool, first) {
		t.Fatalf("content not found in cultural pools: %q", first)
	}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
