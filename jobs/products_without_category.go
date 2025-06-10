package jobs

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"io"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/library/log"
	"reflect"
	"strings"
)

type ProductsWithoutCategory struct {
	Job
}

func (pwc *ProductsWithoutCategory) Run() {
	pwc.handleProducts()
}

func (pwc *ProductsWithoutCategory) handleProducts() {
	products, scrollId := pwc.getProducts("")
	for {
		if len(products) == 0 {
			break
		}

		var productIds []int64
		var i = 0
		for _, product := range products {
			productIds = append(productIds, product.Id)
			i++
		}

		pwc.addToListener(productIds)

		log.GetLogger().Info(fmt.Sprintf("Handled Product: %d", i))
		products, scrollId = pwc.getProducts(scrollId)
	}
}

func (pwc *ProductsWithoutCategory) getProducts(scrollId string) ([]*elasticsearch_entities.Product, string) {
	ctx := context.Background()

	lim := 1000
	search := pwc.Elastic.Scroll("epa").Size(lim)

	var searchResult *elastic.SearchResult
	var err error

	if scrollId != "" {
		searchResult, err = search.ScrollId(scrollId).
			Pretty(true).
			Do(ctx)
	} else {
		q := elastic.NewBoolQuery()

		q.Must(elastic.NewTermQuery("shop_active", true))
		q.Must(elastic.NewTermQuery("active", true))
		q.Must(elastic.NewRangeQuery("stock").From(0).IncludeLower(false))
		q.MustNot(elastic.NewTermQuery("image_passive", true))
		q.MustNot(elastic.NewTermQuery("banned", true))
		q.Must(elastic.NewTermsQuery("evo_category.id", 0))
		searchResult, err = search.Query(q).Pretty(false).Do(ctx)
	}

	if err != nil {
		if err == io.EOF && searchResult != nil {
			pwc.Elastic.ClearScroll(searchResult.ScrollId)
		} else {
			log.GetLogger().Error("elasticsearch error", err)
		}

		return nil, ""
	} else {
		if searchResult.TotalHits() == 0 && searchResult.TotalHits() <= 1000 {
			pwc.Elastic.ClearScroll(searchResult.ScrollId)
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

func (pwc *ProductsWithoutCategory) addToListener(ids []int64) {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := pwc.DB.Exec(sql)

	fmt.Println(fmt.Sprintf("added to product listener %d", res.RowsAffected))
}
