package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"strings"
)

type MoveProductLocks struct {
	Job
}

func (mpl *MoveProductLocks) Run() {
	var minId int64 = 0
	for {
		locks := mpl.getLocks(minId)

		if len(locks) == 0 {
			break
		}

		mpl.addToNewTable(locks)

		fmt.Println(fmt.Sprintf("Min Id: %d", minId))
	}
}

func (mpl *MoveProductLocks) getLocks(minId int64) []entities.EProductLocks {
	var locks []entities.EProductLocks

	mpl.DB.Model(&entities.EProductLocks{}).
		Where("id > ?", minId).
		Limit(1000).
		Order("id asc").
		Scan(&locks)

	return locks
}

func (mpl *MoveProductLocks) addToNewTable(locks []entities.EProductLocks) {
	sql := "INSERT INTO e_product_locks_v2 (`product_id`, `field_name`, `created_at`, `start_date`, `end_date`, `direction`) VALUES "
	var sqlArray []string

	for _, lock := range locks {
		elem := fmt.Sprintf("(%d,'%s','%s','%s','%s', %d)", lock.ProductID, lock.FieldName, lock.CreatedAt, lock.StartDate, lock.EndDate, lock.Direction)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := mpl.DB.Exec(sql)

	fmt.Println(fmt.Sprintf("added to product listener %d", res.RowsAffected))
}
