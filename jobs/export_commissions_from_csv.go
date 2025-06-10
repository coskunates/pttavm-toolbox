package jobs

import (
	"my_toolbox/helpers"
	"my_toolbox/models"
)

type ExportCommissionsFromCsv struct {
	Job
}

func (ecc *ExportCommissionsFromCsv) Run() {
	records, err := helpers.ReadFromCSV("assets/commission/wrong_commission_fix.csv")
	if err != nil {
		panic(err)
	}

	var commissions []models.FixCommission
	for _, record := range records {
		if record[4] == "" {
			record[4] = record[2]
		}
		commissions = append(commissions, models.FixCommission{
			CategoryId:            record[0],
			CategoryName:          record[1],
			DefaultCommissionRate: record[2],
			CommissionRate:        record[3],
			LastRate:              record[4],
			ShopId:                record[5],
		})
	}

	mappedCommissions := make(map[string]models.FixCommission)
	for _, com := range commissions {
		if _, ok := mappedCommissions[com.CategoryId]; !ok {
			mappedCommissions[com.CategoryId] = com
		}
	}

	var data [][]string
	for _, com := range mappedCommissions {
		data = append(data, []string{
			"",
			"",
			com.CategoryId,
			com.CategoryName,
			com.LastRate,
		})
	}

	helpers.Write("assets/commission/fixed_commissions.csv", data)
}
