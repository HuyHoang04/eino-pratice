package main

import "testing"

// runSearch runs search_crm_database with a raw JSON object (email or phone).
func runSearch(t *testing.T, args string) *CRMRecord {
	tl, err := NewSearchCRMDatabaseTool()
	if err != nil {
		t.Fatal(err)
	}
	return mustRun[CRMRecord](t, tl.InvokableRun, args)
}

func TestSearchCRMDatabase(t *testing.T) {
	// Found by email, case-insensitive → returns the full record.
	if rec := runSearch(t, `{"email":"MINH@ACME.IO"}`); !rec.Found || rec.Name != "Minh" || rec.Company != "Acme" || rec.Source != "CRM database" {
		t.Errorf("Minh by email (caps) = %+v, want found Minh/Acme", rec)
	}
	// Found by phone, matched on digits only (separators ignored).
	if rec := runSearch(t, `{"phone":"0909 112 233"}`); !rec.Found || rec.Name != "Lan" || rec.Phone != "0909.112.233" {
		t.Errorf("Lan by phone = %+v, want found Lan with canonical phone", rec)
	}
	// A name is NOT a lookup key (it is not unique): a name-only query never matches.
	if rec := runSearch(t, `{"name":"Lan"}`); rec.Found {
		t.Errorf("name is not a lookup key; Lan by name should not be found, got %+v", rec)
	}
	// Unknown email → found=false (forces the agent to ask).
	if rec := runSearch(t, `{"email":"nobody@nowhere.xyz"}`); rec.Found {
		t.Errorf("unknown email should not be found, got %+v", rec)
	}
	// Unknown phone → found=false.
	if rec := runSearch(t, `{"phone":"0000000000"}`); rec.Found {
		t.Errorf("unknown phone should not be found, got %+v", rec)
	}
	// Empty query → not found, no panic.
	if rec := runSearch(t, `{}`); rec.Found {
		t.Errorf("empty query should not be found, got %+v", rec)
	}
}
