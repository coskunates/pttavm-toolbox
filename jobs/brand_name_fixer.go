package jobs

import (
	"fmt"
	"gorm.io/gorm"
	"my_toolbox/entities"
	"os"
	"strings"
)

type BrandNameFixer struct {
	Job
}

func (bnf *BrandNameFixer) Run() {
	records := bnf.getRecords()
	if len(records) == 0 {
		fmt.Println("bitti")
		os.Exit(0)
	}
	//var data [][]string
	for _, record := range records {
		bnf.update(record)
	}

	fmt.Println("bitti")
	os.Exit(1)

}

func (bnf *BrandNameFixer) getRecords() []entities.EPropertyDetail {
	sql := "SELECT " +
		"* " +
		"FROM e_property_detail " +
		"WHERE property_id = 6564 " +
		"AND (property_value LIKE \"%\\'%\" OR property_value like CONCAT('%', CHAR(0xC2, 0xA0)))"

	var propertyDetailGroups []entities.EPropertyDetail

	bnf.DB.Raw(sql).Scan(&propertyDetailGroups)

	return propertyDetailGroups
}

func (bnf *BrandNameFixer) update(brand entities.EPropertyDetail) {
	propertyValue := brand.PropertyValue
	propertyValue = strings.Replace(propertyValue, "\\'", "'", -1)
	propertyValue = strings.ReplaceAll(propertyValue, "\u00A0", "")

	sql := bnf.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&entities.EPropertyDetail{}).Where("detail_id", brand.DetailID).Update("property_value", propertyValue)
	})

	fmt.Println("SQL:", sql)

	bnf.DB.Model(&entities.EPropertyDetail{}).Where("detail_id", brand.DetailID).Update("property_value", propertyValue)
}
