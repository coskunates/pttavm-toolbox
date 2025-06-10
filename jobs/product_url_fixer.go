package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
)

type ProductUrlFixer struct {
	Job
}

func (puf *ProductUrlFixer) Run() {
	var minProductId int64 = 0
	for {
		products := puf.getProducts(minProductId)

		if len(products) == 0 {
			break
		}

		for _, product := range products {
			if product.ProdottoID > minProductId {
				minProductId = product.ProdottoID
			}

			slug := fmt.Sprintf("%d_%s", product.ProdottoID, helpers.MakeProductSlug(product.ProdottoNome))

			sql1 := fmt.Sprintf("UPDATE e_prodotto_content SET prodotto_url='%s' WHERE prodotto_id = %d AND lang_id = '%s'", slug, product.ProdottoID, product.LangID)
			puf.DB.Exec(sql1)

			log.GetLogger().Info(fmt.Sprintf("Old Url: %s New Url: %s", product.ProdottoURL, slug))
		}
	}

	log.GetLogger().Info("url güncelleme işlemleri tamamlandı")
}

func (puf *ProductUrlFixer) getProducts(minProductId int64) []entities.EProdottoContent {
	sql := fmt.Sprintf("SELECT "+
		"* "+
		"FROM "+
		"e_prodotto_content "+
		"WHERE "+
		"e_prodotto_content.prodotto_id > %d AND "+
		"e_prodotto_content.prodotto_url REGEXP '[^a-zA-Z0-9_-]' ORDER BY prodotto_id ASC LIMIT 1000", minProductId)

	var productContents []entities.EProdottoContent

	puf.DB.Raw(sql).Scan(&productContents)

	return productContents
}
