package main

import "testing"

// runPlan runs recommend_crm_plan with a raw JSON object.
func runPlan(t *testing.T, args string) *PlanRecommendation {
	tl, err := NewRecommendCRMPlanTool()
	if err != nil {
		t.Fatal(err)
	}
	return mustRun[PlanRecommendation](t, tl.InvokableRun, args)
}

func TestRecommendCRMPlan(t *testing.T) {
	// 20 users → Standard (cheapest plan with MaxUsers >= 20).
	if r := runPlan(t, `{"team_size":"20"}`); r.Recommended != "Standard" {
		t.Errorf("team 20: recommended %q, want Standard", r.Recommended)
	}
	// 3 users → Starter.
	if r := runPlan(t, `{"team_size":"3"}`); r.Recommended != "Starter" {
		t.Errorf("team 3: recommended %q, want Starter", r.Recommended)
	}
	// 50 users → Professional.
	if r := runPlan(t, `{"team_size":"50"}`); r.Recommended != "Professional" {
		t.Errorf("team 50: recommended %q, want Professional", r.Recommended)
	}
	// 500 users → exceeds every standard plan → Enterprise (custom).
	if r := runPlan(t, `{"team_size":"500"}`); r.Recommended != "Enterprise" {
		t.Errorf("team 500: recommended %q, want Enterprise", r.Recommended)
	}
	// Unknown team size → no pick, catalog still returned.
	r := runPlan(t, `{}`)
	if r.Recommended != "" || len(r.Plans) != 4 {
		t.Errorf("unknown team size: recommended=%q plans=%d (want empty + 4 plans)", r.Recommended, len(r.Plans))
	}
}

func TestParseTeamSize(t *testing.T) {
	for in, want := range map[string]int{
		"20":        20,
		"~20 người": 20,
		"khoảng 25": 25,
		"20-30":     20, // takes the first number
		"":          0,
		"chưa rõ":   0,
	} {
		if got := parseTeamSize(in); got != want {
			t.Errorf("parseTeamSize(%q) = %d, want %d", in, got, want)
		}
	}
}
