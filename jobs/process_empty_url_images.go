package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"strconv"
	"time"
)

type ProcessEmptyUrlImages struct {
	Job
}

func (peu *ProcessEmptyUrlImages) Run() {
	records, err := helpers.ReadFromCSV("assets/product/empty_url_products.csv")
	if err != nil {
		panic("csv is no readed")
	}

	var productIds []int64
	for _, record := range records {
		productId, _ := strconv.ParseInt(record[0], 10, 64)
		productIds = append(productIds, productId)
	}

	chunks := peu.getChunks(productIds, 100)

	for chunkId, chunk := range chunks {
		productIdBasedImages := peu.getProductImages(chunk)

		for _, productImage := range productIdBasedImages {
			productImage = peu.handleProductImage(productImage)

			peu.updateProductImage(productImage)
		}

		fmt.Println(fmt.Sprintf("chunk %d", chunkId))
	}

	fmt.Println("bitti")
}

func (peu *ProcessEmptyUrlImages) getProductImages(productIds []int64) []mongo_entities.ProductImages {
	filters := bson.M{"product_id": bson.M{"$in": productIds}}

	var productImages []mongo_entities.ProductImages

	cursor, err := peu.Mongo.Database("epttavm").Collection("product_images").Find(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &productImages)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	return productImages
}

func (peu *ProcessEmptyUrlImages) getChunks(records []int64, cs int) [][]int64 {
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

func (peu *ProcessEmptyUrlImages) handleProductImage(productImage mongo_entities.ProductImages) mongo_entities.ProductImages {
	for i := range productImage.Images {
		if productImage.Images[i].Url != "" {
			continue
		}

		bigImagePath := ""
		for _, dimension := range productImage.Images[i].Dimensions {
			if dimension.Size == 592 {
				bigImagePath = dimension.FtpPath
				break
			}
		}

		if bigImagePath == "" {
			fmt.Printf("sikinti var bro Product Id: %d\n", productImage.ProductId)
			continue
		}

		productImage.Images[i].Url = fmt.Sprintf("https://cdn-img.pttavm.com/%s", bigImagePath)
		now := time.Now()
		productImage.Images[i].UpdatedAt = &now
	}

	return productImage
}

func (peu *ProcessEmptyUrlImages) updateProductImage(image mongo_entities.ProductImages) {
	filters := bson.M{}

	if image.Id != primitive.NilObjectID {
		filters["_id"] = image.Id
	}

	opts := options.Update().SetUpsert(true)

	updateRes, err := peu.Mongo.Database("epttavm").Collection("product_images").UpdateOne(context.Background(), filters, bson.M{"$set": &image}, opts)
	if err != nil {
		log.GetLogger().Graylog(true).Error("[DiscountLock] update mongo discount lock error", err)
	}

	if updateRes.ModifiedCount > 0 {
		fmt.Println(fmt.Sprintf("ProductId: %d güncellendi.", image.ProductId))
	} else {
		fmt.Println(fmt.Sprintf("ProductId: %d güncellenEMEdi.", image.ProductId))
	}
}
