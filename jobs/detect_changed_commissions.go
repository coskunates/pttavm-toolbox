package jobs

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"reflect"
	"sync"
)

type DetectChangedCommissions struct {
	Job
}

func (dcc *DetectChangedCommissions) Run() {
	categories := dcc.getCategories()

	chunks := dcc.getChunks(categories, 100)

	for idx, chunk := range chunks {
		wg := &sync.WaitGroup{}

		for _, category := range chunk {
			wg.Add(1)
			go dcc.handleWrongCommissions(wg, category)
		}

		wg.Wait()
		fmt.Println(fmt.Sprintf("%d indexli chunk tamamlandı", idx))
	}

}
func (dcc *DetectChangedCommissions) getCategories() []entities.EvoCategories {
	var categories []entities.EvoCategories
	dcc.DB.Where("id NOT IN (?)", []int{6724, 6725}).Find(&categories)

	return categories
}

func (dcc *DetectChangedCommissions) handleWrongCommissions(wg *sync.WaitGroup, category entities.EvoCategories) {
	defer wg.Done()
	q := elastic.NewBoolQuery()

	q.Must(elastic.NewTermQuery("entity_id", category.ID))
	q.Must(elastic.NewTermQuery("entity_type", "evo_category"))
	q.MustNot(elastic.NewTermQuery("ratio", category.DefaultCommissionRate))

	offset := 0
	limit := 2000
	for {
		searchResult, err := dcc.Elastic.Search().
			Index("commissions").
			Query(q).
			From(offset).Size(limit).
			Pretty(true).
			Do(context.Background())
		if err != nil {
			log.GetLogger().Error("elastic get error", err)
		}

		if len(searchResult.Hits.Hits) <= 0 {
			break
		}

		var data [][]string

		var commission elasticsearch_entities.Commission
		for _, item := range searchResult.Each(reflect.TypeOf(commission)) {
			if c, ok := item.(elasticsearch_entities.Commission); ok {
				createdAt := ""
				if c.CreatedAt != nil {
					createdAt = c.CreatedAt.Format("2006-01-02 15:04:05")
				}

				updatedAt := ""
				if c.UpdatedAt != nil {
					updatedAt = c.UpdatedAt.Format("2006-01-02 15:04:05")
				}

				createdBy := ""
				if c.CreatedBy != "" {
					createdBy = c.CreatedBy
				}

				updatedBy := ""
				if c.UpdatedBy != "" {
					updatedBy = c.UpdatedBy
				}
				row := fmt.Sprintf("%d;%s;%d;%.2f;%d;%s;%s;%s;%s", category.ID, category.Name, category.DefaultCommissionRate, c.Ratio, c.ShopId, createdBy, createdAt, updatedBy, updatedAt)
				data = append(data, []string{row})
			}
		}

		if len(data) > 0 {
			helpers.Write("assets/commission/wrong_commissions.csv", data)
		}

		offset += limit
	}
}

func (dcc *DetectChangedCommissions) getChunks(records []entities.EvoCategories, cs int) [][]entities.EvoCategories {
	var divided [][]entities.EvoCategories

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}
