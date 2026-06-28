package main

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// LeadInfo carries everything the agent has learned about the lead so far.
type LeadInfo struct {
	CustomerName         string   `json:"customer_name,omitempty" jsonschema:"description=Customer name"`
	Phone                string   `json:"phone,omitempty" jsonschema:"description=Phone number"`
	Email                string   `json:"email,omitempty" jsonschema:"description=Email"`
	Company              string   `json:"company,omitempty" jsonschema:"description=Company name"`
	TeamSize             string   `json:"team_size,omitempty" jsonschema:"description=Team size"`
	MainNeed             string   `json:"main_need,omitempty" jsonschema:"description=Main need or the solution the customer is looking for"`
	CurrentState         string   `json:"current_state,omitempty" jsonschema:"description=Current situation or tools currently in use"`
	PainPoint            string   `json:"pain_point,omitempty" jsonschema:"description=Current pain point"`
	DesiredGoal          string   `json:"desired_goal,omitempty" jsonschema:"description=Desired goal"`
	Budget               string   `json:"budget,omitempty" jsonschema:"description=Estimated budget"`
	PreferredContactTime string   `json:"preferred_contact_time,omitempty" jsonschema:"description=Preferred time to be contacted"`
	Notes                string   `json:"notes,omitempty" jsonschema:"description=Additional notes"`
	SkippedFields        []string `json:"skipped_fields,omitempty" jsonschema:"description=Fields the customer refused to provide (e.g. phone or budget) so they are not asked again"`
}

// LeadCheckReport: present/missing fields, next field to ask, sample question, handoff flag.
type LeadCheckReport struct {
	Ready             bool     `json:"ready"`
	ContactResolved   bool     `json:"contact_resolved"`
	PresentMustHave   []string `json:"present_must_have"`
	MissingMustHave   []string `json:"missing_must_have"`
	MissingNiceToHave []string `json:"missing_nice_to_have"`
	SkippedFields     []string `json:"skipped_fields"`
	NextField         string   `json:"next_field"`
	SuggestedQuestion string   `json:"suggested_question"`
	RecommendHandoff  bool     `json:"recommend_handoff"`
}

// Field classification (the deterministic "case thinking"):
//   - Must-have: customer_name, main_need, contact (phone OR email).
//   - Nice-to-have: everything else. Lives here, not in the model.
func NewCheckLeadRequirementsTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"check_lead_requirements",
		"Check whether a sales lead has enough information to be advised. Returns the missing must-have fields, the single next field to ask, a sample question, and whether to hand off to a human consultant. Call this tool every turn before replying to the customer.",
		func(ctx context.Context, input *LeadInfo) (*LeadCheckReport, error) {
			// Skipped set: lower-cased, trimmed keys.
			skipped := make(map[string]bool, len(input.SkippedFields))
			for _, f := range input.SkippedFields {
				skipped[strings.ToLower(strings.TrimSpace(f))] = true
			}

			// A field is present iff its value is non-empty AND not skipped.
			present := func(val, key string) bool {
				return strings.TrimSpace(val) != "" && !skipped[key]
			}

			nameOk := present(input.CustomerName, "customer_name")
			mainNeedOk := present(input.MainNeed, "main_need")
			phoneOk := present(input.Phone, "phone")
			emailOk := present(input.Email, "email")
			contactResolved := phoneOk || emailOk

			report := &LeadCheckReport{
				ContactResolved: contactResolved,
				SkippedFields:   input.SkippedFields,
			}

			// Present must-have labels.
			if nameOk {
				report.PresentMustHave = append(report.PresentMustHave, "customer_name")
			}
			if mainNeedOk {
				report.PresentMustHave = append(report.PresentMustHave, "main_need")
			}
			if contactResolved {
				report.PresentMustHave = append(report.PresentMustHave, "contact")
			}

			// Missing must-have labels, in priority order.
			if !nameOk {
				report.MissingMustHave = append(report.MissingMustHave, "customer_name")
			}
			if !mainNeedOk {
				report.MissingMustHave = append(report.MissingMustHave, "main_need")
			}
			if !contactResolved {
				report.MissingMustHave = append(report.MissingMustHave, "contact")
			}

			// Missing nice-to-have labels, in struct order.
			niceToHaves := []struct{ val, key string }{
				{input.Company, "company"},
				{input.TeamSize, "team_size"},
				{input.CurrentState, "current_state"},
				{input.PainPoint, "pain_point"},
				{input.DesiredGoal, "desired_goal"},
				{input.Budget, "budget"},
				{input.PreferredContactTime, "preferred_contact_time"},
				{input.Notes, "notes"},
			}
			for _, n := range niceToHaves {
				if !present(n.val, n.key) {
					report.MissingNiceToHave = append(report.MissingNiceToHave, n.key)
				}
			}

			// Ready when every must-have is present.
			report.Ready = nameOk && mainNeedOk && contactResolved

			// Next field to ask (priority), with a sample question.
			switch {
			case !nameOk:
				report.NextField = "customer_name"
				report.SuggestedQuestion = "Could you share your name so I can address you properly?"
			case !mainNeedOk:
				report.NextField = "main_need"
				report.SuggestedQuestion = "What specific problem are you looking for a solution to?"
			case !contactResolved:
				report.NextField = "contact"
				report.SuggestedQuestion = "Could you share a phone number or email so I can send materials and recommend a suitable plan?"
			default:
				report.NextField = ""
				report.SuggestedQuestion = ""
			}

			// Handoff when the core need is refused, or no contact channel can be obtained.
			report.RecommendHandoff = skipped["main_need"] || (skipped["phone"] && skipped["email"])

			return report, nil
		},
	)
}
