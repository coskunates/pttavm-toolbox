package jobs

import (
	"fmt"
	"my_toolbox/entities"
)

type DuplicateBrandDetector struct {
	Job
}

func (db *DuplicateBrandDetector) Run() {
	offset := 0
	limit := 100

	for {
		records := db.getRecords(offset, limit)
		if len(records) == 0 {
			fmt.Println("bitti")
			break
		}
		//var data [][]string
		for _, record := range records {
			brand := db.getBrand(record)
			db.deleteOthers(brand)
		}

		//helpers.Write("assets/brand/wrong_brands.csv", data)

		offset += limit
	}

}

func (db *DuplicateBrandDetector) getRecords(offset int, limit int) []entities.EPropertyDetailGroup {
	sql := fmt.Sprintf("SELECT "+
		"property_value, count(detail_id) as total "+
		"FROM e_property_detail "+
		"WHERE property_id = 6564 AND del=0 "+
		"GROUP BY property_value "+
		"HAVING total > 1 "+
		"ORDER BY total DESC LIMIT %d,%d", offset, limit)

	var propertyDetailGroups []entities.EPropertyDetailGroup

	db.DB.Raw(sql).Scan(&propertyDetailGroups)

	return propertyDetailGroups
}

func (db *DuplicateBrandDetector) getBrand(record entities.EPropertyDetailGroup) entities.EPropertyDetail {
	var records entities.EPropertyDetail

	db.DB.Where("property_id", 6564).Where("property_value", record.PropertyValue).Order("detail_id ASC").First(&records)

	return records
}

func (db *DuplicateBrandDetector) deleteOthers(brand entities.EPropertyDetail) {
	sql1 := fmt.Sprintf("UPDATE e_property_detail SET del=0 WHERE detail_id = %d", brand.DetailID)

	db.DB.Exec(sql1)
	sql2 := fmt.Sprintf("UPDATE e_property_detail SET del=1 WHERE property_id = 6564 AND del=0 AND detail_id > %d AND property_value='%s'", brand.DetailID, brand.PropertyValue)

	res := db.DB.Exec(sql2)

	fmt.Println(brand.PropertyValue, res.RowsAffected)
}
