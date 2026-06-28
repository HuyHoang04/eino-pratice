package main

import (
	"encoding/json"
	"testing"
)

// runCheck runs check_lead_requirements with a LeadInfo serialized to JSON.
func runCheck(t *testing.T, info LeadInfo) *LeadCheckReport {
	tl, err := NewCheckLeadRequirementsTool()
	if err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(info)
	return mustRun[LeadCheckReport](t, tl.InvokableRun, string(args))
}

func TestCheckLeadRequirements(t *testing.T) {
	// 1. Nothing known → name is the top-priority gap, not ready.
	r := runCheck(t, LeadInfo{})
	if r.Ready || r.NextField != "customer_name" || !contains(r.MissingMustHave, "main_need") {
		t.Errorf("empty: ready=%v next=%q missing=%v", r.Ready, r.NextField, r.MissingMustHave)
	}

	// 2. Name + need but no contact → contact is the next gap.
	r = runCheck(t, LeadInfo{CustomerName: "Hoang", MainNeed: "CRM cho ban le"})
	if r.Ready || r.NextField != "contact" || !contains(r.PresentMustHave, "customer_name") {
		t.Errorf("no-contact: ready=%v next=%q present=%v", r.Ready, r.NextField, r.PresentMustHave)
	}

	// 3. Phone satisfies the contact requirement → ready.
	r = runCheck(t, LeadInfo{CustomerName: "Hoang", MainNeed: "CRM", Phone: "0909.000.111"})
	if !r.Ready || !r.ContactResolved || r.NextField != "" {
		t.Errorf("phone-only: ready=%v contact=%v next=%q", r.Ready, r.ContactResolved, r.NextField)
	}

	// 4. Email alone also satisfies contact → ready.
	r = runCheck(t, LeadInfo{CustomerName: "Hoang", MainNeed: "CRM", Email: "hoang@x.com"})
	if !r.Ready || !r.ContactResolved {
		t.Errorf("email-only: ready=%v contact=%v", r.Ready, r.ContactResolved)
	}

	// 5. Nice-to-have completeness: when ready, nice-to-have gaps are still reported.
	r = runCheck(t, LeadInfo{CustomerName: "H", MainNeed: "CRM", Phone: "1"})
	if len(r.MissingNiceToHave) != 8 {
		t.Errorf("ready-but-sparse: missing nice-to-have = %d, want 8 (%v)", len(r.MissingNiceToHave), r.MissingNiceToHave)
	}
}

func TestCheckLeadRequirements_RefusalAndHandoff(t *testing.T) {
	// Customer refused phone → pivot to email, no handoff yet.
	r := runCheck(t, LeadInfo{CustomerName: "Tuan", MainNeed: "CRM", SkippedFields: []string{"phone"}})
	if r.NextField != "contact" || r.RecommendHandoff {
		t.Errorf("phone-refused: next=%q handoff=%v (want email next, no handoff)", r.NextField, r.RecommendHandoff)
	}

	// Phone value present but field skipped → must NOT count as present/contact.
	r = runCheck(t, LeadInfo{CustomerName: "T", MainNeed: "CRM", Phone: "0909", SkippedFields: []string{"phone"}})
	if r.ContactResolved {
		t.Errorf("skipped phone should not resolve contact; ContactResolved=%v", r.ContactResolved)
	}

	// Both contact channels refused → handoff.
	r = runCheck(t, LeadInfo{CustomerName: "Tuan", MainNeed: "CRM", SkippedFields: []string{"phone", "email"}})
	if !r.RecommendHandoff || r.Ready {
		t.Errorf("both-contacts-refused: handoff=%v ready=%v (want handoff)", r.RecommendHandoff, r.Ready)
	}

	// Skipped keys are case/whitespace-tolerant.
	r = runCheck(t, LeadInfo{CustomerName: "T", MainNeed: "CRM", Phone: "1", SkippedFields: []string{"  PHONE "}})
	if r.ContactResolved {
		t.Errorf("case-insensitive skip of PHONE should clear contact; ContactResolved=%v", r.ContactResolved)
	}

	// Core need refused → handoff even with contact present.
	r = runCheck(t, LeadInfo{CustomerName: "T", Phone: "1", SkippedFields: []string{"main_need"}})
	if !r.RecommendHandoff {
		t.Errorf("main_need refused should recommend handoff; got %+v", r)
	}
}
