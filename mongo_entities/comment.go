package mongo_entities

import "time"

type Comment struct {
	Id         int       `json:"id" bson:"_id,omitempty"`
	UserId     int       `json:"user_id" bson:"user_id"`
	Comment    string    `json:"comment" bson:"comment"`
	Rating     int32     `json:"rating" bson:"rating"`
	ProductId  int       `json:"product_id" bson:"product_id"`
	IP         string    `json:"ip" bson:"ip"`
	ShopId     int       `json:"shop_id" bson:"shop_id"`
	CategoryId int       `json:"category_id" bson:"category_id"`
	Publish    int       `json:"publish" bson:"publish"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}
