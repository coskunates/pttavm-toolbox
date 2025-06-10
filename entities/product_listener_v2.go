package entities

type ProductListenerV2 struct {
	Id           int    `gorm:"column:id;primary_key"`
	ProductId    int    `gorm:"column:product_id"`
	Action       string `gorm:"column:action"`
	IndexerGroup int    `gorm:"column:indexer_group"`
}

// TableName sets the insert table name for this struct type
func (e *ProductListenerV2) TableName() string {
	return "product_listener_v2"
}
