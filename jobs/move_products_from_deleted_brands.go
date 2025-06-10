package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"strings"
	"sync"
)

type MoveProductsFromDeletedBrands struct {
	Job
}

func (mpf *MoveProductsFromDeletedBrands) Run() {
	duplicateBrands := mpf.getDuplicateBrands()

	for _, duplicateBrand := range duplicateBrands {
		detailIds := strings.Split(duplicateBrand.DetailIds, ",")
		brands := mpf.getBrands(detailIds)

		if len(brands) > 0 {
			var activeBrand entities.EPropertyDetail
			var passiveBrands []entities.EPropertyDetail
			for _, brand := range brands {
				if brand.Del == 0 {
					activeBrand = brand
				} else {
					passiveBrands = append(passiveBrands, brand)
				}
			}

			if activeBrand.DetailID > 0 && len(passiveBrands) > 0 {
				wg := sync.WaitGroup{}
				for _, passiveBrand := range passiveBrands {
					wg.Add(1)
					go mpf.moveProductsTo(&wg, passiveBrand, activeBrand)
				}

				wg.Wait()

				fmt.Println(fmt.Sprintf("%s markası ile ilgili taşıma işlemleri tamamlandı.", duplicateBrand.PropertyValue))
			}
		}
	}
}

func (mpf *MoveProductsFromDeletedBrands) getDuplicateBrands() []entities.EPropertyDetailGroup {
	sql := "SELECT " +
		"property_value, group_concat(detail_id) as detail_ids, count(detail_id) as total " +
		"FROM e_property_detail " +
		"WHERE property_id = 6564 " +
		"GROUP BY property_value " +
		"HAVING total > 1 " +
		"ORDER BY total DESC"

	var propertyDetailGroups []entities.EPropertyDetailGroup

	mpf.DB.Raw(sql).Scan(&propertyDetailGroups)

	return propertyDetailGroups
}

func (mpf *MoveProductsFromDeletedBrands) getBrands(ids []string) []entities.EPropertyDetail {
	var brands []entities.EPropertyDetail
	mpf.DB.Model(entities.EPropertyDetail{}).
		Where("detail_id IN (?)", ids).
		Find(&brands)

	return brands
}

func (mpf *MoveProductsFromDeletedBrands) moveProductsTo(wg *sync.WaitGroup, passiveBrand entities.EPropertyDetail, activeBrand entities.EPropertyDetail) {
	defer wg.Done()

	limit := 1000

	for {
		products := mpf.getProducts(passiveBrand.DetailID, limit)
		if len(products) <= 0 {
			break
		}

		var ids []string
		var productIds []int
		for _, product := range products {
			ids = append(ids, fmt.Sprintf("%d", product.ID))
			productIds = append(productIds, product.ProdottoID)
		}

		updatedPropertyCount := mpf.updateEPropertyProdotto(ids, activeBrand.DetailID, activeBrand.PropertyValue)
		updateProducts := mpf.updateProducts(productIds)
		fmt.Println(fmt.Sprintf("%s - %d -> %d Updated property count: %d - updated product count: %d", passiveBrand.PropertyValue, passiveBrand.DetailID, activeBrand.DetailID, updatedPropertyCount, updateProducts))
	}
}

func (mpf *MoveProductsFromDeletedBrands) getProducts(detailId int, limit int) []entities.EPropertyProdotto {
	var products []entities.EPropertyProdotto
	mpf.DB.Model(entities.EPropertyProdotto{}).
		Where("detail_key", detailId).
		Limit(limit).
		Find(&products)

	return products
}

func (mpf *MoveProductsFromDeletedBrands) updateEPropertyProdotto(ids []string, detailId int, propertyValue string) int64 {
	sql := fmt.Sprintf("UPDATE e_property_prodotto SET detail_key = %d, detail_value = '%s' WHERE id IN (%s)", detailId, propertyValue, strings.Join(ids, ","))
	res := mpf.DB.Exec(sql)

	return res.RowsAffected
}

func (mpf *MoveProductsFromDeletedBrands) updateProducts(ids []int) int64 {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := mpf.DB.Exec(sql)

	return res.RowsAffected
}
