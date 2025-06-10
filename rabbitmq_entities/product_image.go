package rabbitmq_entities

type ProductImage struct {
	ProductId      int64   `json:"product_id"`
	ProductBarcode string  `json:"product_barcode"`
	Images         []Image `json:"images"`
	ShopId         int64   `json:"shop_id"`
	Type           string  `json:"type"`
	ProductName    string  `json:"product_name"`
}

type Image struct {
	Order int32  `json:"order"`
	Url   string `json:"url"`
}
