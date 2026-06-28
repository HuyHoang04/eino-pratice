package main

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CRMSearchInput is the argument schema for the search_crm_database tool.
type CRMSearchInput struct {
	Name  string `json:"name,omitempty" jsonschema:"description=Customer name to look up"`
	Email string `json:"email,omitempty" jsonschema:"description=Customer email to look up"`
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

// NewSearchCRMDatabaseTool looks up a customer by name or email (case-insensitive) to auto-fill fields.
func NewSearchCRMDatabaseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"search_crm_database",
		"Look up an existing customer in the CRM database by name or email. Use this to auto-fill missing fields such as a phone number before asking the customer.",
		func(ctx context.Context, input *CRMSearchInput) (*CRMRecord, error) {
			name := strings.ToLower(strings.TrimSpace(input.Name))
			email := strings.ToLower(strings.TrimSpace(input.Email))
			if name == "" && email == "" {
				return &CRMRecord{Found: false}, nil
			}
			for _, rec := range crmDatabase {
				matchName := name != "" && strings.EqualFold(rec.Name, name)
				matchEmail := email != "" && strings.EqualFold(rec.Email, email)
				if matchName || matchEmail {
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
