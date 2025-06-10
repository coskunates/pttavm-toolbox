package elasticsearch_entities

import "time"

type Commission struct {
	Id         string     `json:"id"`
	Priority   int        `json:"priority"`
	Ratio      float64    `json:"ratio"`
	EntityId   int        `json:"entity_id"`
	EntityType string     `json:"entity_type"`
	ShopId     int        `json:"shop_id"`
	RatioDate  float64    `json:"ratio_date,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedBy  string     `json:"updated_by,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}
