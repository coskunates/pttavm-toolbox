package jobs

import (
	"fmt"
	"my_toolbox/helpers"
	"strconv"
	"strings"
)

type ProductUpdateIndexer struct {
	Job
}

func (pui *ProductUpdateIndexer) Run() {
	records, err := helpers.ReadFromCSV("assets/product/update_products.csv")
	if err != nil {
		panic(err)
	}

	var productIds []int64
	for _, record := range records {
		productId, _ := strconv.ParseInt(record[0], 10, 64)
		productIds = append(productIds, productId)
	}

	chunks := pui.getChunks(productIds, 1000)
	for _, chunk := range chunks {
		updatedProducts := pui.updateProducts(chunk)
		fmt.Println(fmt.Sprintf("Updated product count: %d", updatedProducts))
	}
}

func (pui *ProductUpdateIndexer) getChunks(records []int64, cs int) [][]int64 {
	var divided [][]int64

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}

func (pui *ProductUpdateIndexer) updateProducts(ids []int64) int64 {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := pui.DB.Exec(sql)

	return res.RowsAffected
}
