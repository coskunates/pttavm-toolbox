package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"strconv"
)

type ImageExistsWithError struct {
	Job
}

func (iee *ImageExistsWithError) Run() {
	records, err := helpers.ReadFromCSV("assets/product/no_image_products.csv")
	if err != nil {
		panic(err)
	}

	var productIds []int64
	for _, record := range records {
		productId, _ := strconv.ParseInt(record[0], 10, 64)
		productIds = append(productIds, productId)
	}

	chunks := iee.getChunks(productIds, 1000)

	for chunkId, chunk := range chunks {
		productImages := iee.getProductImages(chunk)

		if len(productImages) == 0 {
			continue
		}

		var data [][]string
		for _, productImage := range productImages {
			data = append(data, []string{fmt.Sprintf("%d", productImage.ProductId)})
		}

		log.GetLogger().Info(fmt.Sprintf("Chunk ID: %d", chunkId))
		helpers.Write("assets/product/image_exists_with_error_products.csv", data)
	}
}

func (iee *ImageExistsWithError) getChunks(records []int64, cs int) [][]int64 {
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

func (iee *ImageExistsWithError) getProductImages(productIds []int64) map[int64]mongo_entities.ProductImageWithError {
	filters := bson.M{"product_id": bson.M{"$in": productIds}}

	var productImages []mongo_entities.ProductImageWithError

	cursor, err := iee.Mongo.Database("epttavm").Collection("product_images_with_error").Find(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &productImages)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	mappedProducts := make(map[int64]mongo_entities.ProductImageWithError)
	for _, productImage := range productImages {
		mappedProducts[productImage.ProductId] = productImage
	}

	return mappedProducts
}
