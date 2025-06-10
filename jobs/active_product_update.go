package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/library/log"
	"sync"
	"time"
)

const (
	productUpdateThreadLimit  = 10000000
	productUpdateProductLimit = 10000
	productPageLimit          = 101
)

var pc int64

type ActiveProductUpdate struct {
	Job
}

func (apu *ActiveProductUpdate) Run() {
	pc = 0
	log.GetLogger().Info("active product updates started")

	wg := sync.WaitGroup{}
	for i := 1; i <= productPageLimit; i++ {
		wg.Add(1)
		go apu.updateActiveProducts(&wg, i)
	}

	wg.Wait()

	log.GetLogger().Info(fmt.Sprintf("active product updates ended. Total Product Count: %d", pc))

	log.GetLogger().Info("bitti")
}

func (apu *ActiveProductUpdate) updateActiveProducts(wg *sync.WaitGroup, i int) {
	defer wg.Done()

	minId := (i - 1) * productUpdateThreadLimit
	maxId := i * productUpdateThreadLimit

	for j := minId; j <= maxId; j += productUpdateProductLimit {
		a := apu.DB.Model(&entities.EProdotto{}).Where("prodotto_id > ?", j).
			Where("prodotto_id <= ?", j+productUpdateProductLimit).
			Where("prodotto_attivo = ?", "1").
			Update("last_mod", time.Now())

		pc += a.RowsAffected

		log.GetLogger().Info(fmt.Sprintf("Thread: %d - Min ID: %d - Total Updated: %d", i, j, pc))
	}
}
