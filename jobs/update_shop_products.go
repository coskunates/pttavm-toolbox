package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"strconv"
	"strings"
)

type UpdateShopProducts struct {
	Job

	// Job-specific parameters only
	ShopId        int64  `param:"shop_id" help:"Specific shop ID to process (0 = process all from CSV)" default:"0"`
	FilePath      string `param:"file_path" help:"CSV file path containing shop IDs" default:"assets/shop_ids.csv"`
	ProductStatus string `param:"product_status" help:"Product status filter: active, passive, or empty for all" default:"active"`
}

func (usp *UpdateShopProducts) Run() {
	// Default değerler
	if usp.Limit == 0 {
		usp.Limit = 1000
	}
	if usp.FilePath == "" {
		usp.FilePath = "assets/shop_ids.csv"
	}

	if usp.ProductStatus == "" {
		usp.ProductStatus = "active"
	}

	fmt.Printf("=== Update Shop Products Job Started ===\n")
	fmt.Printf("Parameters: shop_id=%d, limit=%d, file_path=%s, product_status=%s, dry_run=%t, verbose=%t\n",
		usp.ShopId, usp.Limit, usp.FilePath, usp.ProductStatus, usp.DryRun, usp.Verbose)

	if usp.DryRun {
		fmt.Println("🔍 DRY RUN MODE - No actual changes will be made")
	}

	// Eğer specific shop_id verilmişse sadece onu işle
	if usp.ShopId > 0 {
		fmt.Printf("Processing single shop: %d\n", usp.ShopId)
		usp.updateShopProducts(int(usp.ShopId), usp.Limit, usp.DryRun)
	} else {
		// CSV'den oku (mevcut davranış)
		fmt.Printf("Reading shop IDs from file: %s\n", usp.FilePath)
		records, err := helpers.ReadFromCSV(usp.FilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Found %d shops to process\n", len(records))
		for i, record := range records {
			shopId, _ := strconv.Atoi(record[0])
			if usp.Verbose {
				fmt.Printf("Processing shop %d/%d: %d\n", i+1, len(records), shopId)
			}
			usp.updateShopProducts(shopId, usp.Limit, usp.DryRun)
		}
	}

	fmt.Println("=== Job Completed ===")
}

func (usp *UpdateShopProducts) updateShopProducts(id int, limit int, dryRun bool) {
	maxProductId := 0
	totalProcessed := 0

	for {
		products := usp.getProducts(id, maxProductId, limit)
		if len(products) <= 0 {
			break
		}

		var productIds []int
		for _, product := range products {
			productIds = append(productIds, product.ProdottoID)
			if product.ProdottoID > maxProductId {
				maxProductId = product.ProdottoID
			}
		}

		if dryRun {
			displayIds := productIds
			if len(productIds) > 5 {
				displayIds = productIds[:5]
			}
			fmt.Printf("  [DRY RUN] Would update %d products (First 5 IDs: %v)\n", len(productIds), displayIds)
		} else {
			updateProducts := usp.updateProducts(productIds)
			if usp.Verbose {
				fmt.Printf("  Updated product count: %d\n", updateProducts)
			}
		}

		totalProcessed += len(productIds)
	}

	if usp.Verbose {
		fmt.Printf("Shop %d: Total products processed: %d\n", id, totalProcessed)
	}
}

func (usp *UpdateShopProducts) getProducts(shopId, maxProductId int, limit int) []entities.EProdotto {
	var products []entities.EProdotto

	query := usp.DB.Model(entities.EProdotto{}).
		Where("shop_id", shopId).
		Where("prodotto_id > ?", maxProductId).
		Order("prodotto_id ASC").
		Limit(limit)

	if usp.ProductStatus == "active" {
		query = query.Where("prodotto_attivo", "1")
	} else if usp.ProductStatus == "passive" {
		query = query.Where("prodotto_attivo", "0")
	}

	query.Find(&products)

	return products
}

func (usp *UpdateShopProducts) updateProducts(ids []int) int64 {
	sql := "INSERT IGNORE INTO product_listener_v2 (`product_id`, `action`, `created_at`, `resource`, `priority`) VALUES "
	var sqlArray []string

	for _, id := range ids {
		elem := fmt.Sprintf("(%d,'%s',NOW(),'%s',%d)", id, "update", "manuel_update", 1)
		sqlArray = append(sqlArray, elem)
	}

	sql += strings.Join(sqlArray, ",")

	res := usp.DB.Exec(sql)

	return res.RowsAffected
}
