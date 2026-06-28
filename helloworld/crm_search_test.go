package main

import "testing"

// runSearch runs search_crm_database with a raw JSON object (both fields optional).
func runSearch(t *testing.T, args string) *CRMRecord {
	tl, err := NewSearchCRMDatabaseTool()
	if err != nil {
		t.Fatal(err)
	}
	return mustRun[CRMRecord](t, tl.InvokableRun, args)
}

func TestSearchCRMDatabase(t *testing.T) {
	// Found by name → phone auto-filled.
	if rec := runSearch(t, `{"name":"Lan"}`); !rec.Found || rec.Phone != "0909.112.233" || rec.Source != "CRM database" {
		t.Errorf("Lan by name = %+v, want found with phone 0909.112.233", rec)
	}
	// Found by email, case-insensitive.
	if rec := runSearch(t, `{"email":"MINH@ACME.IO"}`); !rec.Found || rec.Name != "Minh" || rec.Company != "Acme" {
		t.Errorf("Minh by email (caps) = %+v, want found Minh/Acme", rec)
	}
	// Not seeded → found=false (forces the agent to ask).
	if rec := runSearch(t, `{"name":"Hoang"}`); rec.Found {
		t.Errorf("Hoang should not be seeded, got %+v", rec)
	}
	// Missing both keys → not found, no panic.
	if rec := runSearch(t, `{}`); rec.Found {
		t.Errorf("empty query should not be found, got %+v", rec)
	}
	// Whitespace-only name trimmed and treated as empty.
	if rec := runSearch(t, `{"name":"   "}`); rec.Found {
		t.Errorf("whitespace name should not be found, got %+v", rec)
	}
}
