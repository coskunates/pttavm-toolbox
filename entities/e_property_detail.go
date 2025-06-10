package entities

type EPropertyDetail struct {
	Del             int    `gorm:"column:del"`
	DetailID        int    `gorm:"column:detail_id;primary_key"`
	LangID          string `gorm:"column:lang_id"`
	PropertyID      int    `gorm:"column:property_id"`
	PropertyOrderID int    `gorm:"column:property_order_id"`
	PropertyValue   string `gorm:"column:property_value"`
}

// TableName sets the insert table name for this struct type
func (e *EPropertyDetail) TableName() string {
	return "e_property_detail"
}
