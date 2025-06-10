package jobs

import (
	"my_toolbox/helpers"
)

type ExportNoImageProductsShops struct {
	Job
}

func (eip *ExportNoImageProductsShops) Run() {
	records, err := helpers.ReadFromCSV("assets/product/no_image_products.csv")
	if err != nil {
		panic(err)
	}

	shopsUnique := make(map[string]bool)
	for _, record := range records {
		if _, ok := shopsUnique[record[2]]; !ok {
			shopsUnique[record[2]] = true
		}
	}

	var data [][]string
	for shopName, _ := range shopsUnique {
		data = append(data, []string{shopName})
	}

	helpers.Write("assets/product/no_image_products_shops.csv", data)
}
