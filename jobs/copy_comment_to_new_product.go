package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
)

type CopyCommentToNewProduct struct {
	Job
}

func (ccp *CopyCommentToNewProduct) Run() {
	comments := ccp.getComments(519432436)

	for _, comment := range comments {
		ccp.addToNewProduct(1097040398, comment)
	}

}

func (ccp *CopyCommentToNewProduct) getComments(productId int64) []mongo_entities.Comment {
	filters := bson.M{"product_id": productId}

	var comments []mongo_entities.Comment

	cursor, err := ccp.ReviewMongo.Database("epttavm").Collection("comments").Find(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &comments)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	return comments
}

func (ccp *CopyCommentToNewProduct) addToNewProduct(productId int, comment mongo_entities.Comment) {
	nextSeq, err := ccp.getNextSequence()
	if err != nil {
		return
	}

	newComment := mongo_entities.Comment{
		Id:         nextSeq,
		UserId:     comment.UserId,
		Comment:    comment.Comment,
		Rating:     comment.Rating,
		ProductId:  productId,
		IP:         comment.IP,
		ShopId:     comment.ShopId,
		CategoryId: comment.CategoryId,
		Publish:    comment.Publish,
		CreatedAt:  comment.CreatedAt,
	}

	_, iErr := ccp.ReviewMongo.Database("epttavm").Collection("comments").InsertOne(context.Background(), newComment)

	if iErr != nil {
		fmt.Println(iErr.Error())
		return
	}
}

func (ccp *CopyCommentToNewProduct) getNextSequence() (int, error) {
	commentSequence := mongo_entities.CommentSequence{}
	filter := bson.M{"_id": "comment_id"} // Replace with your unique key
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := ccp.ReviewMongo.Database("epttavm").Collection("sequence").FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&commentSequence)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, err
		} else {
			return 0, err
		}
	}

	return commentSequence.Sequence, nil
}
