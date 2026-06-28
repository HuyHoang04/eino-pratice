package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CRMPlan is one product plan in the catalog.
type CRMPlan struct {
	Name             string   `json:"name"`
	PriceVNDPerMonth int      `json:"price_vnd_per_month"`
	MaxUsers         int      `json:"max_users"` // 0 = unlimited / custom
	Features         []string `json:"features"`
}

// crmPlans is the SAMPLE product catalog. Replace names/prices/features here
// with real values — this is the only place plan data lives.
var crmPlans = []CRMPlan{
	{Name: "Starter", PriceVNDPerMonth: 990000, MaxUsers: 5, Features: []string{
		"Basic customer management",
		"Simple sales pipeline",
		"Overview reports",
	}},

	{Name: "Standard", PriceVNDPerMonth: 2900000, MaxUsers: 25, Features: []string{
		"Customer management with interaction history",
		"Team-based access control",
		"Reports & dashboards",
		"Basic workflow automation",
	}},

	{Name: "Professional", PriceVNDPerMonth: 5900000, MaxUsers: 100, Features: []string{
		"Everything in the Standard plan",
		"Marketing automation",
		"API & integrations",
		"Advanced reporting",
	}},

	{Name: "Enterprise", PriceVNDPerMonth: 0, MaxUsers: 0, Features: []string{
		"Customizable to business needs",
		"Dedicated SLA & support",
		"On-premises or private cloud deployment",
	}},
}

// PlanRecommendationInput is the argument schema for recommend_crm_plan.
type PlanRecommendationInput struct {
	TeamSize string `json:"team_size,omitempty" jsonschema:"description=Team size"`
}

// PlanRecommendation is returned by recommend_crm_plan.
type PlanRecommendation struct {
	Recommended string    `json:"recommended"`
	Reason      string    `json:"reason"`
	Plans       []CRMPlan `json:"plans"`
}

// NewRecommendCRMPlanTool builds recommend_crm_plan: picks the cheapest plan
// that fits the team size. Unknown team size -> no pick, just the catalog.
func NewRecommendCRMPlanTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"recommend_crm_plan",
		"Recommend a CRM plan based on team size. Returns the recommended plan with price and features plus the full catalog. Call this before suggesting any plan or price. Never invent plan names prices or features.",
		func(ctx context.Context, input *PlanRecommendationInput) (*PlanRecommendation, error) {
			n := parseTeamSize(input.TeamSize)
			out := &PlanRecommendation{Plans: crmPlans}
			if n <= 0 {
				out.Reason = "Team size unknown. Ask the customer for team size before recommending."
				return out, nil
			}
			for i := range crmPlans {
				p := &crmPlans[i]
				if p.MaxUsers > 0 && p.MaxUsers >= n {
					out.Recommended = p.Name
					out.Reason = fmt.Sprintf("Team of %d fits %s (up to %d users) at %d VND/month.", n, p.Name, p.MaxUsers, p.PriceVNDPerMonth)
					return out, nil
				}
			}
			last := &crmPlans[len(crmPlans)-1]
			out.Recommended = last.Name
			out.Reason = fmt.Sprintf("Team of %d exceeds every standard plan; recommend %s (custom quote).", n, last.Name)
			return out, nil
		},
	)
}

// parseTeamSize extracts the first integer in s (e.g. "~20 người" -> 20); 0 if none.
func parseTeamSize(s string) int {
	n := 0
	for _, r := range strings.TrimSpace(s) {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		} else if n > 0 {
			break
		}
	}
	return n
}
