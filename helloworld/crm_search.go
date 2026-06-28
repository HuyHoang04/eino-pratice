package main

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CRMSearchInput is the argument schema for the search_crm_database tool.
// Lookup is by a unique identifier only — email or phone. A name is deliberately
// NOT accepted: several customers may share the same name, so it can never be
// relied on to identify the right person.
type CRMSearchInput struct {
	Email string `json:"email,omitempty" jsonschema:"description=Customer email to look up"`
	Phone string `json:"phone,omitempty" jsonschema:"description=Customer phone number to look up"`
}

// CRMRecord is a customer record, returned by search_crm_database.
type CRMRecord struct {
	Found        bool   `json:"found"`
	Name         string `json:"name,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
	Company      string `json:"company,omitempty"`
	TeamSize     string `json:"team_size,omitempty"`
	CurrentState string `json:"current_state,omitempty"`
	Source       string `json:"source,omitempty"`
}

// crmDatabase is the mock CRM. Hoàng is NOT seeded (forces asking); Lan and Minh are (auto-fill demo).
var crmDatabase = []CRMRecord{
	{
		Name:         "Lan",
		Email:        "lan@globex.com",
		Phone:        "0909.112.233",
		Company:      "Globex JSC",
		TeamSize:     "~15",
		CurrentState: "đang dùng Zoho CRM cũ, muốn đổi",
	},
	{
		Name:         "Minh",
		Email:        "minh@acme.io",
		Phone:        "0988.765.432",
		Company:      "Acme",
		TeamSize:     "~40",
		CurrentState: "chưa dùng CRM, quản lý qua Google Sheets",
	},
}

// NewSearchCRMDatabaseTool looks up a customer by email or phone (case-insensitive;
// phone matched on digits only). Both uniquely identify one person, unlike a name
// which several customers may share, so the lookup always resolves to the right
// record. Used to auto-fill fields such as company, team size, or phone.
func NewSearchCRMDatabaseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"search_crm_database",
		"Look up an existing customer in the CRM database by their email or phone — these uniquely identify one person, unlike a name which may be shared by several customers. Use this to auto-fill fields such as the customer's company, team size, or phone before asking them.",
		func(ctx context.Context, input *CRMSearchInput) (*CRMRecord, error) {
			email := strings.ToLower(strings.TrimSpace(input.Email))
			phone := normalizePhone(input.Phone)
			if email == "" && phone == "" {
				return &CRMRecord{Found: false}, nil
			}
			for _, rec := range crmDatabase {
				if (email != "" && strings.EqualFold(rec.Email, email)) ||
					(phone != "" && normalizePhone(rec.Phone) == phone) {
					matched := rec
					matched.Found = true
					matched.Source = "CRM database"
					return &matched, nil
				}
			}
			return &CRMRecord{Found: false}, nil
		},
	)
}

// normalizePhone keeps digits only so "0909.112.233", "0909 112 233" and
// "0909112233" all match the same record.
func normalizePhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
