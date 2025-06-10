package entities

import (
	"database/sql"
	"time"
)

type EvoCategories struct {
	CreatedAt             time.Time       `gorm:"column:created_at"`
	CreatedBy             sql.NullString  `gorm:"column:created_by"`
	DefaultCommissionRate int             `gorm:"column:default_commission_rate"`
	DeletedAt             time.Time       `gorm:"column:deleted_at"`
	DeletedBy             sql.NullString  `gorm:"column:deleted_by"`
	Description           string          `gorm:"column:description"`
	ID                    int             `gorm:"column:id;primary_key"`
	IsActive              int             `gorm:"column:is_active"`
	IsLandingPage         int             `gorm:"column:is_landing_page"`
	Keywords              sql.NullString  `gorm:"column:keywords"`
	MaxInstallment        int             `gorm:"column:max_installment"`
	MaxPrice              sql.NullFloat64 `gorm:"column:max_price"`
	MetaDescription       string          `gorm:"column:meta_description"`
	MetaKeywords          string          `gorm:"column:meta_keywords"`
	MetaTitle             string          `gorm:"column:meta_title"`
	MinInstallmentPrice   float64         `gorm:"column:min_installment_price"`
	Name                  string          `gorm:"column:name"`
	OnMenu                int             `gorm:"column:on_menu"`
	ParentID              int             `gorm:"column:parent_id"`
	R2sCategoryID         sql.NullString  `gorm:"column:r2s_category_id"`
	RelatedCategories     sql.NullString  `gorm:"column:related_categories"`
	Title                 string          `gorm:"column:title"`
	UpdatedAt             time.Time       `gorm:"column:updated_at"`
	UpdatedBy             sql.NullString  `gorm:"column:updated_by"`
	URL                   string          `gorm:"column:url"`
}

// TableName sets the insert table name for this struct type
func (e *EvoCategories) TableName() string {
	return "evo_categories"
}
