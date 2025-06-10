package entities

import "time"

type EProductLocksV2 struct {
	ID        int64     `gorm:"column:id;primary_key" json:"id" bson:"id"`
	ProductID int64     `gorm:"column:product_id" json:"product_id" bson:"product_id"`
	FieldName string    `gorm:"column:field_name" json:"field_name" bson:"field_name"`
	StartDate time.Time `gorm:"column:start_date" json:"start_date" bson:"start_date"`
	EndDate   time.Time `gorm:"column:end_date" json:"end_date" bson:"end_date"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" bson:"created_at"`
	Direction int64     `gorm:"column:direction" json:"direction" bson:"direction"`
}

// TableName sets the insert table name for this struct type
func (e *EProductLocksV2) TableName() string {
	return "e_product_locks_v2"
}
