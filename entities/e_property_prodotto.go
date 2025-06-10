package entities

type EPropertyProdotto struct {
	DetailID    int    `gorm:"column:detail_id"`
	DetailKey   int    `gorm:"column:detail_key"`
	DetailValue string `gorm:"column:detail_value"`
	ID          int    `gorm:"column:id;primary_key"`
	ProdottoID  int    `gorm:"column:prodotto_id"`
}

// TableName sets the insert table name for this struct type
func (e *EPropertyProdotto) TableName() string {
	return "e_property_prodotto"
}
