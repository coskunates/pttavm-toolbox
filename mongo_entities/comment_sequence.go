package mongo_entities

type CommentSequence struct {
	Id       string `json:"id" bson:"_id,omitempty"`
	Sequence int    `json:"seq" bson:"seq"`
}
