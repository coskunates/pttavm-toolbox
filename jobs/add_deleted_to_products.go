package jobs

import (
	"fmt"
	"gorm.io/gorm"
	"my_toolbox/entities"
	"strings"
	"time"
)

type AddDeletedToProducts struct {
	Job
}

func (adp *AddDeletedToProducts) Run() {
	shopIds := []int{366698, 115100}

	for _, shopId := range shopIds {
		adp.updateShopProducts(shopId)
	}
}

func (adp *AddDeletedToProducts) updateShopProducts(id int) {
	limit := 1000

	for {
		products := adp.getProducts(id, limit)
		if len(products) <= 0 {
			break
		}

		for _, product := range products {
			if !strings.Contains(product.UrunBarkod, "@Deleted") {
				product.UrunBarkod = "@Deleted-" + product.UrunBarkod
			}

			adp.updateProduct(product)
		}

		fmt.Println(fmt.Sprintf("updated product count: %d", len(products)))
	}
}

func (adp *AddDeletedToProducts) getProducts(shopId, limit int) []entities.EProdotto {
	sql := fmt.Sprintf("SELECT ep.* "+
		"FROM e_prodotto ep "+
		"LEFT JOIN epttavm.e_prodotto_extra epe ON ep.prodotto_id = epe.prodotto_id "+
		"WHERE "+
		"ep.shop_id = %d "+ // Burada 'ep' eklenmesi gerekiyor
		"AND epe.out_of_sale = 1 "+ // Burada 'ep' eklenmesi gerekiyor
		"AND ep.urun_barkod NOT LIKE '@Deleted%%' "+
		"LIMIT %d", shopId, limit)

	var products []entities.EProdotto

	adp.DB.Raw(sql).Scan(&products)

	return products
}

func (adp *AddDeletedToProducts) updateProduct(product entities.EProdotto) {
	sql := adp.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&entities.EProdotto{}).Where("prodotto_id", product.ProdottoID).Updates(map[string]interface{}{
			"last_mod":    time.Now(),
			"urun_barkod": product.UrunBarkod,
		})
	})

	fmt.Println("SQL:", sql)

	adp.DB.Model(&entities.EProdotto{}).Where("prodotto_id", product.ProdottoID).Updates(map[string]interface{}{
		"last_mod":    time.Now(),
		"urun_barkod": product.UrunBarkod,
	})
}
