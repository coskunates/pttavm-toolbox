package mongo_entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProductImages struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductId int64              `json:"product_id" bson:"product_id,omitempty"`
	Images    []ProductImage     `json:"images" bson:"images,omitempty"`
}

type ProductImage struct {
	Order      int8        `json:"order" bson:"order,omitempty"`
	Checksum   string      `json:"checksum" bson:"checksum,omitempty"`
	Url        string      `json:"url" bson:"url,omitempty"`
	Dimensions []Dimension `json:"dimensions" bson:"dimensions,omitempty"`
	UpdatedAt  *time.Time  `json:"updated_at" bson:"updated_at,omitempty"`
}

type Dimension struct {
	Size    int    `json:"size" bson:"size,omitempty"`
	Name    string `json:"name" bson:"name,omitempty"`
	FtpPath string `json:"ftp_path" bson:"ftp_path,omitempty"`
}
