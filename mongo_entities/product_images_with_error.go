package mongo_entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProductImageWithError struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductId      int64              `json:"product_id" bson:"product_id"`
	ShopId         int64              `json:"shop_id" bson:"shop_id"`
	ProductBarcode string             `json:"product_barcode" bson:"product_barcode"`
	Infos          []Info             `json:"infos" bson:"infos"`
}

type Info struct {
	Order      int32     `json:"order" bson:"order"`
	Url        string    `json:"url" bson:"url"`
	ErrorInfo  string    `json:"error_info" bson:"error_info"`
	OccurredAt time.Time `json:"occurred_at" bson:"occurred_at"`
}
