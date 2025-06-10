package jobs

import (
	"fmt"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
)

type ChunkNoBrandProducts struct {
	Job
}

func (enp *ChunkNoBrandProducts) Run() {
	records, err := helpers.ReadFromCSV("assets/product/no_brand_products.csv")
	if err != nil {
		panic("read file error")
	}

	count := 0
	delta := 1

	var data [][]string
	for idx, record := range records {
		if count == 500000 {
			fileName := fmt.Sprintf("assets/product/no_brand_products_%d.csv", delta)
			helpers.Write(fileName, data)
			delta++
			count = 0
			data = [][]string{}
		}

		data = append(data, []string{record[0], record[1], record[2]})
		count++
		log.GetLogger().Info(fmt.Sprintf("Idx: %d", idx))
	}

	if len(data) > 0 {
		fileName := fmt.Sprintf("assets/product/no_brand_products_%d.csv", delta)
		helpers.Write(fileName, data)
	}

	log.GetLogger().Info("bitti")
}
