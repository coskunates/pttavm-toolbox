package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"sync"
	"time"
)

var (
	EmptyUrlImages chan [][]string
)

type DetectEmptyUrlImages struct {
	Job
}

func (deu *DetectEmptyUrlImages) Run() {
	EmptyUrlImages = make(chan [][]string)

	go deu.listenChannel()

	var maxId int64 = 1128524045
	stepSize := helpers.GetStepSize(maxId, 10)

	var i int64
	wg := sync.WaitGroup{}
	for i = 0; i <= maxId; i += stepSize {
		maxProductId := i + stepSize
		if i+stepSize > maxId {
			maxProductId = maxId
		}

		wg.Add(1)
		go deu.handleChunks(&wg, i, maxProductId)
	}

	wg.Wait()

	log.GetLogger().Info("bekliyoruz")
	time.Sleep(60 * time.Second)

	log.GetLogger().Info("bitti")
}

func (deu *DetectEmptyUrlImages) handleChunks(wg *sync.WaitGroup, minProductId int64, maxProductId int64) {
	defer wg.Done()

	var steps []helpers.Step
	steps = helpers.GetSteps(minProductId, maxProductId, 100000)
	stepChunks := helpers.GetStepChunks(steps, 30)

	for stepChunkIdx, stepChunk := range stepChunks {
		feedWaitGroup := sync.WaitGroup{}
		for _, step := range stepChunk {
			feedWaitGroup.Add(1)
			go deu.handleChunk(&feedWaitGroup, step.Minimum, step.Maximum)
		}

		log.GetLogger().Info(fmt.Sprintf("StepChunkIdx: %d is completed", stepChunkIdx))
		feedWaitGroup.Wait()
	}
}

func (deu *DetectEmptyUrlImages) handleChunk(wg *sync.WaitGroup, minimum int64, maximum int64) {
	defer wg.Done()

	var limit int64 = 10000
	for {
		if minimum > maximum {
			break
		}

		log.GetLogger().Info(fmt.Sprintf("Min Product ID: %dMax Product ID: %d", minimum, minimum+limit))
		productIds := deu.getEmptyUrlImages(minimum, minimum+limit)
		if productIds == nil {
			minimum = minimum + limit
			continue
		}
		var data [][]string
		for _, productId := range productIds {
			data = append(data, []string{fmt.Sprintf("%d", productId)})
		}

		EmptyUrlImages <- data

		minimum = minimum + limit
	}
}

func (deu *DetectEmptyUrlImages) getEmptyUrlImages(minimum int64, maximum int64) []int64 {
	filters := bson.M{}

	filters["product_id"] = bson.M{"$gt": minimum, "$lte": maximum}
	filters["images.url"] = bson.M{"$exists": false}

	var productImages []mongo_entities.ProductImages

	cursor, err := deu.Mongo.Database("epttavm").Collection("product_images").Find(context.Background(), filters)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &productImages)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	var productIds []int64

	for _, productImage := range productImages {
		productIds = append(productIds, productImage.ProductId)
	}

	return productIds
}

func (deu *DetectEmptyUrlImages) listenChannel() {
	go func() {
		for {
			select {
			case products := <-EmptyUrlImages:
				helpers.Write("assets/product/empty_url_products.csv", products)
			}
		}
	}()
}
