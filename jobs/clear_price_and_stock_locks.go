package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"strings"
)

type ClearPriceAndStockLocks struct {
	Job
}

func (cl *ClearPriceAndStockLocks) Run() {
	var minId int64 = 0
	for {
		locks := cl.getLocks(minId)

		if len(locks) == 0 {
			break
		}

		var lockIds []int64
		var productIds []int64
		for _, lock := range locks {
			lockIds = append(lockIds, lock.ID)
			productIds = append(productIds, lock.ProductID)

			if lock.ID > minId {
				minId = lock.ID
			}
		}

		if len(lockIds) > 0 {
			cl.deleteLocks(lockIds)
		}

		if len(productIds) > 0 {
			cl.addToListener(productIds)
		}

		fmt.Println(fmt.Sprintf("Min Id: %d", minId))
	}
}

func (cl *ClearPriceAndStockLocks) getLocks(minId int64) []entities.EProductLocks {
	var locks []entities.EProductLocks

	cl.DB.Model(&entities.EProductLocks{}).
		Where("id > ?", minId).
		Where("field_name IN (?)", []string{"price", "stock"}).
		Limit(1000).
		Order("id asc").
		Scan(&locks)

	return locks
}

func (cl *ClearPriceAndStockLocks) deleteLocks(lockIds []int64) {
	result := cl.DB.Model(&entities.EProductLocks{}).
		Where("id  IN (?)", lockIds).
		Delete(entities.EProductLocks{})
	pc = pc + result.RowsAffected
}

func (cl *ClearPriceAndStockLocks) addToListener(ids []int64) {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := cl.DB.Exec(sql)

	fmt.Println(fmt.Sprintf("added to product listener %d", res.RowsAffected))
}
