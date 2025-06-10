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

type CommissionFixer struct {
	Job
}

var totalUpdatedCommissionCount int

func (cf *CommissionFixer) Run() {
	categories, categoryIds := cf.getCategories()

	shopLimit := 100
	skip := 0
	for {
		shops := cf.getShops(skip, shopLimit)

		if len(shops) <= 0 {
			break
		}

		wg := sync.WaitGroup{}
		for _, shop := range shops {
			wg.Add(1)
			go cf.checkCommissions(&wg, shop, categories, categoryIds)
		}
		wg.Wait()

		log.GetLogger().Info(fmt.Sprintf("Total Count: %d", totalUpdatedCommissionCount))
		skip += shopLimit
	}
}
func (cf *CommissionFixer) getCategories() (map[int]entities.EvoCategories, []interface{}) {
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

func (cf *CommissionFixer) getShops(skip int, limit int) []entities.EShop {
	log.GetLogger().Info(fmt.Sprintf("Skip: %d Limit: %d", skip, limit))
	var shops []entities.EShop
	cf.DB.
		Limit(limit).
		Offset(skip).
		Find(&shops)

	return shops
}

func (cf *CommissionFixer) checkCommissions(wg *sync.WaitGroup, shop entities.EShop, categories map[int]entities.EvoCategories, ids []interface{}) {
	defer wg.Done()

	commissions := cf.getCommissions(shop, ids)

	var updateCategoryIds [][]string
	bulkRequest := cf.Elastic.Bulk()
	for _, category := range categories {
		commissionId := fmt.Sprintf("%d_evo_category_%d", shop.ShopID, category.ID)
		if _, ok := commissions[commissionId]; !ok {
			updateCategoryIds = append(updateCategoryIds, []string{fmt.Sprintf("%d", shop.ShopID), fmt.Sprintf("%d", category.ID)})
			now := time.Now()
			coms := elasticsearch_entities.Commission{
				Id:         commissionId,
				Priority:   80,
				Ratio:      float64(category.DefaultCommissionRate),
				EntityId:   category.ID,
				EntityType: "evo_category",
				ShopId:     shop.ShopID,
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

		log.GetLogger().Info(fmt.Sprintf("%d mağazasına %d adet komisyon eklendi", shop.ShopID, len(updateCategoryIds)))
	}
}

func (cf *CommissionFixer) getCommissions(shop entities.EShop, ids []interface{}) map[string]elasticsearch_entities.Commission {
	q := elastic.NewBoolQuery()

	q.Must(elastic.NewTermsQuery("entity_id", ids...))
	q.Must(elastic.NewTermQuery("entity_type", "evo_category"))
	q.Must(elastic.NewTermQuery("shop_id", shop.ShopID))

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
