package jobs

import (
	"fmt"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"strconv"
)

type AssignBarcode struct {
	ProductId int64
	Barcode   string
}

type AssignBarcodes struct {
	Job
}

func (ab *AssignBarcodes) Run() {
	barcodesTemp, err := helpers.ReadFromCSV("assets/product/assign_barcodes.csv")
	if err != nil {
		log.GetLogger().Error("read csv error", err)
	}

	var barcodes []AssignBarcode
	for _, barcode := range barcodesTemp {
		productId, _ := strconv.ParseInt(barcode[0], 10, 64)

		barcodes = append(barcodes, AssignBarcode{productId, barcode[1]})
	}

	ab.updateBarcodes(barcodes)

	fmt.Println("bitti")
}

func (ab *AssignBarcodes) updateBarcodes(barcodes []AssignBarcode) {
	for _, barcode := range barcodes {
		updateSql := fmt.Sprintf("UPDATE e_prodotto SET urun_barkod='%s', prodotto_codice='%s' WHERE prodotto_id = %d", barcode.Barcode, barcode.Barcode, barcode.ProductId)

		tx := ab.DB.Exec(updateSql)
		log.GetLogger().Info(fmt.Sprintf("Affected Rows: %d SQL: %s", tx.RowsAffected, updateSql))
	}
}
