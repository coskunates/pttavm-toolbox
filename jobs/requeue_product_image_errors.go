package jobs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"my_toolbox/rabbitmq_entities"
	"sync"
)

type RequeueProductImageErrors struct {
	Job
}

func (rpi *RequeueProductImageErrors) Run() {

	errorImages := rpi.getErrorImages()

	chunks := rpi.getChunks(errorImages, 100)

	wg := &sync.WaitGroup{}
	for _, chunk := range chunks {
		wg.Add(1)
		go rpi.addToRabbitMQ(wg, chunk)
	}

	wg.Wait()
}

func (rpi *RequeueProductImageErrors) getErrorImages() []mongo_entities.ProductImageWithError {
	filters := bson.M{"shop_id": bson.M{"$in": []int{17844}}}

	var productImages []mongo_entities.ProductImageWithError

	cursor, err := rpi.Mongo.Database("epttavm").Collection("product_images_with_error").Find(context.Background(), filters)
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

func (rpi *RequeueProductImageErrors) getChunks(records []mongo_entities.ProductImageWithError, cs int) [][]mongo_entities.ProductImageWithError {
	var divided [][]mongo_entities.ProductImageWithError

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}

func (rpi *RequeueProductImageErrors) addToRabbitMQ(wg *sync.WaitGroup, chunk []mongo_entities.ProductImageWithError) {
	defer wg.Done()

	headers := make(amqp.Table)
	headers["x-delay"] = 0
	headers["x-try"] = 5

	for _, errorImage := range chunk {
		productImage := rabbitmq_entities.ProductImage{
			ProductId:      errorImage.ProductId,
			ProductBarcode: errorImage.ProductBarcode,
			ShopId:         errorImage.ShopId,
			Type:           "update",
		}

		var images []rabbitmq_entities.Image
		for _, image := range errorImage.Infos {
			images = append(images, rabbitmq_entities.Image{
				Order: image.Order,
				Url:   image.Url,
			})
		}

		productImage.Images = images
		messageBody, _ := json.Marshal(productImage)

		err := rpi.PttAvmRabbitMQ.Publish(headers, &amqp.Delivery{
			Exchange:   "images.add",
			RoutingKey: "",
			Body:       []byte(base64.StdEncoding.EncodeToString(messageBody)),
		})

		if err != nil {
			log.GetLogger().Error(
				fmt.Sprintf("%s images add requeue error", "toolbox"),
				err,
			)
		}
	}
}
