package models

import "time"

type RuleFilter struct {
	Field  string   `json:"field"`
	Values []string `json:"values"`
}

type RuleValidityPeriod struct {
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type RuleValidity struct {
	Period RuleValidityPeriod `json:"period"`
}
type CreateRuleRequest struct {
	Id             string             `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Type           string             `json:"type"`
	Metadata       *map[string]string `json:"metadata"`
	TermInDays     int                `json:"term_in_days"`
	MerchantID     *int32             `json:"merchant_id"`
	CommissionRate float64            `json:"commission_rate"`
	RuleType       string             `json:"rule_type"`
	Filters        []RuleFilter       `json:"filters"`
	Validity       RuleValidity       `json:"validity"`
	IsActive       bool               `json:"is_active"`
	Status         string             `json:"status"`
}

// Cause provides structure for validation errors.
type Cause struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// RuleResponse defines the standard API response structure.
type RuleResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    CreateRuleRequest `json:"data,omitempty"`
	Error   *APIError         `json:"error,omitempty"` // Embed error details directly
}

// APIError defines the structure for error details in the response.
type APIError struct {
	Code      int         `json:"code,omitempty"`      // Internal error code or category
	Title     string      `json:"title"`               // Short, user-friendly error title
	Detail    string      `json:"detail"`              // Detailed error message for developers/logs
	Causes    []Cause     `json:"causes,omitempty"`    // For validation errors
	Conflicts interface{} `json:"conflicts,omitempty"` // For conflict errors
}
