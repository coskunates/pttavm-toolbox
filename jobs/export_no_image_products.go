package jobs

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"io"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"reflect"
	"sync"
	"time"
)

var (
	NoImageProductChan chan [][]string
)

type ExportNoImageProducts struct {
	Job
}

func (eip *ExportNoImageProducts) Run() {
	NoImageProductChan = make(chan [][]string)

	go eip.listenChannel()

	var maxId int64 = 1060078571
	stepSize := helpers.GetStepSize(maxId, 10)

	var i int64
	wg := sync.WaitGroup{}
	for i = 0; i <= maxId; i += stepSize {
		maxProductId := i + stepSize
		if i+stepSize > maxId {
			maxProductId = maxId
		}

		wg.Add(1)
		go eip.handleChunks(&wg, i, maxProductId)
	}

	wg.Wait()

	log.GetLogger().Info("bekliyoruz")
	time.Sleep(60 * time.Second)

	log.GetLogger().Info("bitti")
}

func (eip *ExportNoImageProducts) handleChunks(wg *sync.WaitGroup, minProductId int64, maxProductId int64) {
	defer wg.Done()

	var steps []helpers.Step
	steps = helpers.GetSteps(minProductId, maxProductId, 1000000)
	stepChunks := helpers.GetStepChunks(steps, 30)

	for stepChunkIdx, stepChunk := range stepChunks {
		feedWaitGroup := sync.WaitGroup{}
		for _, step := range stepChunk {
			feedWaitGroup.Add(1)
			go eip.handleChunk(&feedWaitGroup, step.Minimum, step.Maximum)
		}

		log.GetLogger().Info(fmt.Sprintf("StepChunkIdx: %d is completed", stepChunkIdx))
		feedWaitGroup.Wait()
	}
}

func (eip *ExportNoImageProducts) handleChunk(wg *sync.WaitGroup, minimum int64, maximum int64) {
	defer wg.Done()

	products, scrollId := eip.getNoImageProducts(minimum, maximum, "")
	for {
		if len(products) == 0 {
			break
		}

		var data [][]string
		for _, product := range products {
			data = append(data, []string{fmt.Sprintf("%d", product.Id), product.Name, product.Shop.Name})

			if minimum < product.Id {
				minimum = product.Id
			}
		}

		NoImageProductChan <- data

		log.GetLogger().Info(fmt.Sprintf("Min Product ID: %dMax Product ID: %d", minimum, maximum))

		products, scrollId = eip.getNoImageProducts(minimum, maximum, scrollId)
	}
}

func (eip *ExportNoImageProducts) getNoImageProducts(minProductId, maxProductId int64, scrollId string) ([]*elasticsearch_entities.Product, string) {
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("shop_active", true))
	q.Must(elastic.NewRangeQuery("stock").Gte(1))
	q.Must(elastic.NewRangeQuery("id").Gt(minProductId))
	q.Must(elastic.NewRangeQuery("id").Lte(maxProductId))
	q.Must(elastic.NewTermQuery("image_passive", true))
	q.Must(elastic.NewRangeQuery("evo_category.id").From(0).IncludeLower(false))
	q.MustNot(elastic.NewTermQuery("banned", true))

	ctx := context.Background()

	lim := 1000
	search := eip.Elastic.Scroll("epa").Size(lim)

	var searchResult *elastic.SearchResult
	var err error

	if scrollId != "" {
		searchResult, err = search.ScrollId(scrollId).
			Pretty(true).
			Do(ctx)
	} else {
		searchResult, err = search.Query(q).Pretty(false).Do(ctx)
	}

	if err != nil {
		if err == io.EOF && searchResult != nil {
			eip.Elastic.ClearScroll(searchResult.ScrollId)
		} else {
			log.GetLogger().Error("elasticsearch error", err)
		}

		return nil, ""
	} else {
		if searchResult.TotalHits() == 0 && searchResult.TotalHits() <= 1000 {
			eip.Elastic.ClearScroll(searchResult.ScrollId)
		}

		var products []*elasticsearch_entities.Product
		var product elasticsearch_entities.Product

		for _, item := range searchResult.Each(reflect.TypeOf(product)) {
			if p, ok := item.(elasticsearch_entities.Product); ok {
				products = append(products, &p)
			}
		}

		return products, searchResult.ScrollId
	}
}

func (eip *ExportNoImageProducts) listenChannel() {
	go func() {
		for {
			select {
			case products := <-NoImageProductChan:
				helpers.Write("assets/product/no_image_products.csv", products)
			}
		}
	}()
}
