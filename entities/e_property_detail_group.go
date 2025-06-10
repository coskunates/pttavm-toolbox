package entities

type EPropertyDetailGroup struct {
	PropertyValue string `gorm:"column:property_value"`
	DetailIds     string `gorm:"column:detail_ids"`
	Total         int    `gorm:"column:total"`
}

// TableName sets the insert table name for this struct type
func (e *EPropertyDetailGroup) TableName() string {
	return "e_property_detail"
}
