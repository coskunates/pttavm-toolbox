package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/models"
)

type UpdateCategoryCommissionOnDb struct {
	Job
}

func (ucc *UpdateCategoryCommissionOnDb) Run() {
	records, err := helpers.ReadFromCSV("assets/commission/fixed_commissions.csv")
	if err != nil {
		panic(err)
	}

	var commissions []models.FixCommission
	for _, record := range records {
		if record[4] == "" {
			record[4] = record[2]
		}
		commissions = append(commissions, models.FixCommission{
			CategoryId:            record[2],
			CategoryName:          record[3],
			DefaultCommissionRate: "",
			CommissionRate:        "",
			LastRate:              record[4],
			ShopId:                "",
		})
	}

	for _, comm := range commissions {
		res := ucc.DB.Model(&entities.EvoCategories{}).Where("id", comm.CategoryId).
			Update("default_commission_rate", comm.LastRate)

		fmt.Println(comm.CategoryId, comm.CategoryName, comm.LastRate, res.RowsAffected)
	}
}
