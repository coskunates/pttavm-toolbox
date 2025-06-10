package jobs

import (
	"fmt"
	"my_toolbox/entities"
)

type ProductListenerNotExistedProductCleaner struct {
	Job
}

func (plc *ProductListenerNotExistedProductCleaner) Run() {
	var minId = 0
	for {
		products := plc.getProducts(minId)
		if len(products) == 0 {
			fmt.Println("bitti")
			break
		}

		var ids []int
		for _, product := range products {
			ids = append(ids, product.Id)
			if product.Id > minId {
				minId = product.Id
			}
		}

		plc.deleteFromProductListener(ids)
		fmt.Println(pc)
	}

}

func (plc *ProductListenerNotExistedProductCleaner) getProducts(minId int) []entities.ProductListenerV2 {
	sql := fmt.Sprintf("select "+
		"pl.* "+
		"from product_listener_v2 pl "+
		"WHERE NOT EXISTS (SELECT 1 FROM e_prodotto ep WHERE ep.prodotto_id = pl.product_id) "+
		"AND pl.id > %d ORDER BY pl.id ASC limit 100", minId)

	var result []entities.ProductListenerV2
	plc.DB.Raw(sql).Scan(&result)

	return result
}

func (plc *ProductListenerNotExistedProductCleaner) deleteFromProductListener(ids []int) {
	result := plc.DB.Model(&entities.ProductListenerV2{}).Where("id IN (?)", ids).Delete(entities.ProductListenerV2{})
	pc = pc + result.RowsAffected
}
