package jobs

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/entities"
	"my_toolbox/library/log"
	"reflect"
	"sync"
	"time"
)

type CommissionFixForShops struct {
	Job
}

func (cf *CommissionFixForShops) Run() {
	categories, categoryIds := cf.getCategories()
	shopIds := []int64{575486, 803469, 895509, 876649, 511788, 884909, 892949, 901309, 388928, 519201, 4835, 17153}

	wg := sync.WaitGroup{}
	for _, shopId := range shopIds {
		wg.Add(1)
		cf.checkCommissions(&wg, shopId, categories, categoryIds)
	}

	wg.Wait()

	log.GetLogger().Info(fmt.Sprintf("Total Count: %d", totalUpdatedCommissionCount))
}
func (cf *CommissionFixForShops) getCategories() (map[int]entities.EvoCategories, []interface{}) {
	var categories []entities.EvoCategories
	cf.DB.Where("id NOT IN (?)", []int{6724, 6725}).Find(&categories)

	var categoryIds []interface{}
	categoryByIds := make(map[int]entities.EvoCategories)
	for _, category := range categories {
		categoryIds = append(categoryIds, category.ID)
		categoryByIds[category.ID] = category
	}

	return categoryByIds, categoryIds
}

func (cf *CommissionFixForShops) checkCommissions(wg *sync.WaitGroup, shopId int64, categories map[int]entities.EvoCategories, ids []interface{}) {
	defer wg.Done()

	commissions := cf.getCommissions(shopId, ids)

	var updateCategoryIds [][]string
	bulkRequest := cf.Elastic.Bulk()
	for _, category := range categories {
		commissionId := fmt.Sprintf("%d_evo_category_%d", shopId, category.ID)
		if _, ok := commissions[commissionId]; !ok {
			updateCategoryIds = append(updateCategoryIds, []string{fmt.Sprintf("%d", shopId), fmt.Sprintf("%d", category.ID)})
			now := time.Now()
			coms := elasticsearch_entities.Commission{
				Id:         commissionId,
				Priority:   80,
				Ratio:      float64(category.DefaultCommissionRate),
				EntityId:   category.ID,
				EntityType: "evo_category",
				ShopId:     int(shopId),
				CreatedBy:  "system",
				CreatedAt:  &now,
			}

			req := elastic.NewBulkUpdateRequest().Index("commissions").Id(commissionId).Doc(coms).DocAsUpsert(true)
			bulkRequest.Add(req)
			totalUpdatedCommissionCount++
		}
	}

	if bulkRequest.NumberOfActions() > 0 {
		_, err := bulkRequest.Do(context.Background())
		if err != nil {
			log.GetLogger().Error("bulk request error", err)
		}

		log.GetLogger().Info(fmt.Sprintf("%d mağazasına %d adet komisyon eklendi", shopId, len(updateCategoryIds)))
	}
}

func (cf *CommissionFixForShops) getCommissions(shopId int64, ids []interface{}) map[string]elasticsearch_entities.Commission {
	q := elastic.NewBoolQuery()

	q.Must(elastic.NewTermsQuery("entity_id", ids...))
	q.Must(elastic.NewTermQuery("entity_type", "evo_category"))
	q.Must(elastic.NewTermQuery("shop_id", shopId))

	searchResult, err := cf.Elastic.Search().
		Index("commissions").
		Query(q).
		From(0).Size(len(ids)).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		log.GetLogger().Error("elastic get error", err)
	}

	commissions := make(map[string]elasticsearch_entities.Commission)
	var commission elasticsearch_entities.Commission

	for _, item := range searchResult.Each(reflect.TypeOf(commission)) {
		if c, ok := item.(elasticsearch_entities.Commission); ok {
			commissionId := fmt.Sprintf("%d_evo_category_%d", c.ShopId, c.EntityId)
			commissions[commissionId] = c
		}
	}

	return commissions
}
