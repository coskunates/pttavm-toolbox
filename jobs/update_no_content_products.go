package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"strings"
)

type UpdateNoContentProducts struct {
	Job
}

func (ucp *UpdateNoContentProducts) Run() {
	minProductId := 943681825

	for {
		products := ucp.getProducts(minProductId)

		if len(products) == 0 {
			break
		}

		var productIds []int64
		for _, product := range products {
			if product.ProdottoID > minProductId {
				minProductId = product.ProdottoID
			}

			productIds = append(productIds, int64(product.ProdottoID))
		}

		ucp.addToListener(productIds)

		fmt.Println(productIds)
	}
}

func (ucp *UpdateNoContentProducts) getProducts(minId int) []entities.EProdotto {
	sql := fmt.Sprintf("SELECT "+
		"prodotto_id "+
		"FROM e_prodotto "+
		"WHERE "+
		"prodotto_id > %d and "+
		"prodotto_id not in (select prodotto_id from e_prodotto_content) "+
		"ORDER BY prodotto_id asc "+
		"LIMIT 100", minId)

	var result []entities.EProdotto
	ucp.DB.Raw(sql).Scan(&result)

	return result
}

func (ucp *UpdateNoContentProducts) addToListener(ids []int64) {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := ucp.DB.Exec(sql)

	fmt.Println(fmt.Sprintf("added to product listener %d", res.RowsAffected))
}
