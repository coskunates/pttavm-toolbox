package jobs

import (
	"fmt"
)

type PropertyUpdateSql struct {
	Job
}

func (pus *PropertyUpdateSql) Run() {
	sql := "UPDATE e_property_detail SET del = 0 WHERE detail_id = 105101;" +
		"UPDATE e_property_detail SET del = 1 WHERE detail_id = 605215;" +
		"UPDATE e_property_detail SET del = 1 WHERE detail_id = 545457;" +
		"UPDATE e_property_detail SET del = 0 WHERE detail_id = 105042;" +
		"UPDATE e_property_detail SET del = 0 WHERE detail_id = 94488;" +
		"UPDATE e_property_detail SET del = 1 WHERE detail_id = 183551;" +
		"UPDATE e_property_detail SET del = 0 WHERE detail_id = 97712;" +
		"UPDATE e_property_detail SET del = 1 WHERE detail_id = 585819;" +
		"UPDATE e_property_detail SET del = 0 WHERE detail_id = 98275;" +
		"UPDATE e_property_detail SET del = 1 WHERE detail_id = 527376;"

	res := pus.DB.Exec(sql)
	fmt.Println(res.RowsAffected)
}
