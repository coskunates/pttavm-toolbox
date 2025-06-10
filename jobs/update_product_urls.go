package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"strconv"
	"sync"
)

type UpdateProductUrls struct {
	Job
}

func (usp *UpdateProductUrls) Run() {
	var maxId int64 = 1104134646
	stepSize := helpers.GetStepSize(maxId, 10)

	var i int64
	wg := sync.WaitGroup{}
	for i = 41236744; i <= maxId; i += stepSize {
		maxProductId := i + stepSize
		if i+stepSize > maxId {
			maxProductId = maxId
		}

		wg.Add(1)
		go usp.handleChunks(&wg, i, maxProductId)
	}

	wg.Wait()

	log.GetLogger().Info("bitti")
}

func (usp *UpdateProductUrls) handleChunks(wg *sync.WaitGroup, minProductId int64, maxProductId int64) {
	defer wg.Done()

	var steps []helpers.Step
	steps = helpers.GetSteps(minProductId, maxProductId, 100000)
	stepChunks := helpers.GetStepChunks(steps, 30)

	for stepChunkIdx, stepChunk := range stepChunks {
		feedWaitGroup := sync.WaitGroup{}
		for _, step := range stepChunk {
			feedWaitGroup.Add(1)
			go usp.handleChunk(&feedWaitGroup, step.Minimum, step.Maximum)
		}

		log.GetLogger().Info(fmt.Sprintf("Step %d is completed", stepChunkIdx))
		feedWaitGroup.Wait()
	}
}

func (usp *UpdateProductUrls) handleChunk(wg *sync.WaitGroup, minimum int64, maximum int64) {
	defer wg.Done()

	minProductId := minimum
	maxProductId := minProductId + 1000

	for {
		products := usp.getProducts(minProductId, maxProductId)

		for _, product := range products {
			productName := strconv.Itoa(int(product.ProdottoID)) + "_" + product.ProdottoNome
			slug := helpers.MakeProductSlug(productName)
			sql1 := fmt.Sprintf("UPDATE e_prodotto_content SET prodotto_url='%s' WHERE prodotto_id = %d AND lang_id = '%s'", slug, product.ProdottoID, product.LangID)
			usp.DB.Exec(sql1)
			//fmt.Println(product.ProdottoID, product.ProdottoNome, slug)
		}

		//log.GetLogger().Info(fmt.Sprintf("Min Product ID: %d Max Product ID: %d", minProductId, maxProductId))

		minProductId = maxProductId
		maxProductId = maxProductId + 1000
		if minProductId > maximum {
			break
		}
	}
}

func (usp *UpdateProductUrls) getProducts(minimum int64, maximum int64) []entities.EProdottoContent {
	sql := fmt.Sprintf("SELECT * "+
		"FROM e_prodotto_content "+
		"LEFT JOIN e_prodotto ON e_prodotto_content.prodotto_id = e_prodotto.prodotto_id "+
		"WHERE "+
		"e_prodotto_content.prodotto_url REGEXP '[çÇğĞıİöÖşŞüÜ]' "+
		"AND e_prodotto_content.prodotto_id > %d AND e_prodotto_content.prodotto_id < %d ", minimum, maximum)

	var propertyDetailGroups []entities.EProdottoContent

	usp.DB.Raw(sql).Scan(&propertyDetailGroups)

	return propertyDetailGroups
}
