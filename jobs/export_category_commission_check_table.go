package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/models"
	"strconv"
)

type ExportCategoryCommissionCheckTable struct {
	Job
}

func (ecc *ExportCategoryCommissionCheckTable) Run() {
	records, err := helpers.ReadFromCSV("assets/commission/category_commissions.csv")
	if err != nil {
		panic(err)
	}

	var categoryCommissionTable []models.CategoryCommissionTable
	for _, record := range records {
		categoryId, _ := strconv.ParseInt(record[2], 10, 64)
		commissionRate, _ := strconv.ParseFloat(record[4], 64)
		categoryCommissionTable = append(categoryCommissionTable, models.CategoryCommissionTable{
			CategoryId:     categoryId,
			CommissionRate: commissionRate,
		})
	}

	categories := ecc.getCategories()

	var data [][]string
	for _, ccs := range categoryCommissionTable {
		if _, ok := categories[ccs.CategoryId]; !ok {
			fmt.Println("yohtır lo")
		} else {
			category := categories[ccs.CategoryId]
			if float64(category.DefaultCommissionRate) != ccs.CommissionRate {
				data = append(data, []string{fmt.Sprintf("%d", ccs.CategoryId), category.Name, fmt.Sprintf("%v", ccs.CommissionRate), fmt.Sprintf("%v", category.DefaultCommissionRate)})
			}
		}
	}

	helpers.Write("assets/commission/last_commission.csv", data)

	fmt.Println(categories)

}

func (ecc *ExportCategoryCommissionCheckTable) getCategories() map[int64]entities.EvoCategories {
	var categories []entities.EvoCategories
	ecc.DB.Where("id NOT IN (?)", []int{6724, 6725}).Find(&categories)

	cs := make(map[int64]entities.EvoCategories)
	for _, category := range categories {
		cs[int64(category.ID)] = category
	}

	return cs
}
