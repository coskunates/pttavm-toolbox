package jobs

import (
	"fmt"
	"my_toolbox/helpers"
	"slices"
	"strconv"
)

type ExportNoImageProductsSorted struct {
	Job
}

func (eip *ExportNoImageProductsSorted) Run() {
	records, err := helpers.ReadFromCSV("assets/product/no_image_products.csv")
	if err != nil {
		panic(err)
	}

	var productIds []int64
	for _, record := range records {
		productId, _ := strconv.ParseInt(record[0], 10, 64)
		productIds = append(productIds, productId)
	}

	slices.Sort(productIds)

	var data [][]string
	for _, productId := range productIds {
		data = append(data, []string{fmt.Sprintf("%d", productId)})
	}

	helpers.Write("assets/product/no_image_products_sorted.csv", data)
}
