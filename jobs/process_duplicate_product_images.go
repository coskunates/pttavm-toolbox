package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"strconv"
)

type ProcessDuplicateProductImages struct {
	Job
}

func (pdi *ProcessDuplicateProductImages) Run() {
	records, err := helpers.ReadFromCSV("assets/product/duplicate_product_images.csv")
	if err != nil {
		panic("csv is no readed")
	}

	var productIds []int64
	for _, record := range records {
		productId, _ := strconv.ParseInt(record[0], 10, 64)
		productIds = append(productIds, productId)
	}

	chunks := pdi.getChunks(productIds, 50)

	for chunkId, chunk := range chunks {
		productIdBasedImages := pdi.getProductImages(chunk)

		var deleteProductImages []primitive.ObjectID
		for _, productImages := range productIdBasedImages {
			var lastId primitive.ObjectID
			for _, image := range productImages {
				if lastId.IsZero() {
					lastId = image.Id
				} else if image.Id.Hex() > lastId.Hex() {
					lastId = image.Id
				}
			}

			for _, image := range productImages {
				if image.Id != lastId {
					deleteProductImages = append(deleteProductImages, image.Id)
				}
			}
		}

		if len(deleteProductImages) > 0 {
			pdi.deleteProductImages(deleteProductImages)
		}

		fmt.Println(fmt.Sprintf("chunk %d", chunkId))
	}
}

func (pdi *ProcessDuplicateProductImages) getProductImages(productIds []int64) map[int64][]mongo_entities.ProductImages {
	filters := bson.M{"product_id": bson.M{"$in": productIds}}

	var productImages []mongo_entities.ProductImages

	cursor, err := pdi.Mongo.Database("epttavm").Collection("product_images").Find(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &productImages)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	mappedProducts := make(map[int64][]mongo_entities.ProductImages)
	for _, productImage := range productImages {
		mappedProducts[productImage.ProductId] = append(mappedProducts[productImage.ProductId], productImage)
	}

	return mappedProducts
}

func (pdi *ProcessDuplicateProductImages) getChunks(records []int64, cs int) [][]int64 {
	var divided [][]int64

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}

func (pdi *ProcessDuplicateProductImages) deleteProductImages(ids []primitive.ObjectID) {
	filters := bson.M{}

	if len(ids) <= 0 {
		return
	}

	filters["_id"] = bson.M{"$in": ids}

	delRes, err := pdi.Mongo.Database("epttavm").Collection("product_images").DeleteMany(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	fmt.Println(delRes.DeletedCount)
}
