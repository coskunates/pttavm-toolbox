package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/library/log"
	"sync"
	"time"
)

type PassiveProductUpdate struct {
	Job
}

func (ppu *PassiveProductUpdate) Run() {
	pc = 0
	log.GetLogger().Info("passive product updates started")

	wg := sync.WaitGroup{}
	for i := 1; i <= productPageLimit; i++ {
		wg.Add(1)
		go ppu.updatePassiveProducts(&wg, i)
	}

	wg.Wait()

	log.GetLogger().Info(fmt.Sprintf("passive product updates ended. Total Product Count: %d", pc))

	log.GetLogger().Info("bitti")
}

func (ppu *PassiveProductUpdate) updatePassiveProducts(wg *sync.WaitGroup, i int) {
	defer wg.Done()

	minId := (i - 1) * productUpdateThreadLimit
	maxId := i * productUpdateThreadLimit

	for j := minId; j <= maxId; j += productUpdateProductLimit {
		var products []entities.EProdotto
		ppu.DB.Model(&entities.EProdotto{}).
			Joins("left join e_prodotto_extra on e_prodotto.prodotto_id = e_prodotto_extra.prodotto_id").
			Where("e_prodotto.prodotto_id > ?", j).
			Where("e_prodotto.prodotto_id <= ?", j+productUpdateProductLimit).
			Where("e_prodotto.prodotto_attivo = ?", "0").
			Where("e_prodotto_extra.out_of_sale = ? OR e_prodotto_extra.out_of_sale is null", 0).
			Scan(&products)

		if len(products) > 0 {
			var productIds []int
			for _, prod := range products {
				productIds = append(productIds, prod.ProdottoID)
			}

			c := ppu.DB.Model(&entities.EProdotto{}).Where("prodotto_id IN (?)", productIds).
				Update("last_mod", time.Now())

			pc += c.RowsAffected
		}

		log.GetLogger().Info(fmt.Sprintf("Thread: %d - Min ID: %d - Total Updated: %d", i, j, pc))
	}
}
