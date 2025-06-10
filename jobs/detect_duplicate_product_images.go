package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"sync"
	"time"
)

var (
	DuplicatedProductImages chan [][]string
)

type DetectDuplicateProductImages struct {
	Job
}

func (ddp *DetectDuplicateProductImages) Run() {
	DuplicatedProductImages = make(chan [][]string)

	go ddp.listenChannel()

	var maxId int64 = 1069982931
	stepSize := helpers.GetStepSize(maxId, 10)

	var i int64
	wg := sync.WaitGroup{}
	for i = 621038; i <= maxId; i += stepSize {
		maxProductId := i + stepSize
		if i+stepSize > maxId {
			maxProductId = maxId
		}

		wg.Add(1)
		go ddp.handleChunks(&wg, i, maxProductId)
	}

	wg.Wait()

	log.GetLogger().Info("bekliyoruz")
	time.Sleep(60 * time.Second)

	log.GetLogger().Info("bitti")
}

func (ddp *DetectDuplicateProductImages) handleChunks(wg *sync.WaitGroup, minProductId int64, maxProductId int64) {
	defer wg.Done()

	var steps []helpers.Step
	steps = helpers.GetSteps(minProductId, maxProductId, 100000)
	stepChunks := helpers.GetStepChunks(steps, 30)

	for stepChunkIdx, stepChunk := range stepChunks {
		feedWaitGroup := sync.WaitGroup{}
		for _, step := range stepChunk {
			feedWaitGroup.Add(1)
			go ddp.handleChunk(&feedWaitGroup, step.Minimum, step.Maximum)
		}

		log.GetLogger().Info(fmt.Sprintf("StepChunkIdx: %d is completed", stepChunkIdx))
		feedWaitGroup.Wait()
	}
}

func (ddp *DetectDuplicateProductImages) handleChunk(wg *sync.WaitGroup, minimum int64, maximum int64) {
	defer wg.Done()

	var limit int64 = 10000
	for {
		if minimum > maximum {
			break
		}
		productIds := ddp.getDuplicateProductIds(minimum, minimum+limit)
		if productIds == nil {
			minimum = minimum + limit
			continue
		}
		var data [][]string
		for _, productId := range productIds {
			data = append(data, []string{fmt.Sprintf("%d", productId)})
		}

		DuplicatedProductImages <- data

		minimum = minimum + limit

		log.GetLogger().Info(fmt.Sprintf("Min Product ID: %dMax Product ID: %d", minimum, maximum))
	}
}

func (ddp *DetectDuplicateProductImages) getDuplicateProductIds(minimum int64, maximum int64) []int64 {
	collection := ddp.Mongo.Database("epttavm").Collection("product_images")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "product_id", Value: bson.D{{Key: "$gt", Value: minimum}, {Key: "$lte", Value: maximum}}}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$product_id"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$match", Value: bson.D{{Key: "count", Value: bson.D{{Key: "$gt", Value: 1}}}}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.GetLogger().Info(err.Error())
	}
	defer cursor.Close(context.Background())

	var productIds []int64
	for cursor.Next(context.Background()) {
		var result struct {
			ProductID      int64 `bson:"_id"`
			DuplicateCount int   `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.GetLogger().Info(err.Error())
		}
		productIds = append(productIds, result.ProductID)
	}
	if err := cursor.Err(); err != nil {
		log.GetLogger().Info(err.Error())
	}

	return productIds
}

func (ddp *DetectDuplicateProductImages) listenChannel() {
	go func() {
		for {
			select {
			case products := <-DuplicatedProductImages:
				helpers.Write("assets/product/duplicate_product_images.csv", products)
			}
		}
	}()
}
