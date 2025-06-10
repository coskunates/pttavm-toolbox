package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"my_toolbox/helpers"
	"strconv"
)

// ElasticsearchResponse represents the main response structure
type ElasticsearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore interface{}   `json:"max_score"`
		Hits     []interface{} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		TopShops struct {
			DocCountErrorUpperBound int          `json:"doc_count_error_upper_bound"`
			SumOtherDocCount        int          `json:"sum_other_doc_count"`
			Buckets                 []ShopBucket `json:"buckets"`
		} `json:"top_shops"`
	} `json:"aggregations"`
}

// ShopBucket represents each shop in the aggregation buckets
type ShopBucket struct {
	Key      int64 `json:"key"`       // Shop ID
	DocCount int   `json:"doc_count"` // Product count
	ShopInfo struct {
		Hits struct {
			Total struct {
				Value    int    `json:"value"`
				Relation string `json:"relation"`
			} `json:"total"`
			MaxScore float64 `json:"max_score"`
			Hits     []struct {
				Index  string  `json:"_index"`
				Type   string  `json:"_type"`
				ID     string  `json:"_id"`
				Score  float64 `json:"_score"`
				Source struct {
					Shop struct {
						Name string `json:"name"`
						URL  string `json:"url"`
					} `json:"shop"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	} `json:"shop_info"`
}

// ExportShopAggregation job struct
type ExportShopAggregation struct {
	Job
}

// Run executes the job
func (esa *ExportShopAggregation) Run() {
	// Read the JSON response file
	jsonData, err := ioutil.ReadFile("assets/product/response.json")
	if err != nil {
		log.Fatalf("JSON dosyası okunamadı: %v", err)
	}

	// Parse the JSON response
	var response ElasticsearchResponse
	err = json.Unmarshal(jsonData, &response)
	if err != nil {
		log.Fatalf("JSON parse edilemedi: %v", err)
	}

	// Extract shop data from buckets
	var csvData [][]string

	// Add CSV header
	csvData = append(csvData, []string{
		"Shop ID",
		"Shop Name",
		"Shop URL",
		"Product Count",
	})

	// Process each shop bucket
	for _, bucket := range response.Aggregations.TopShops.Buckets {
		shopID := strconv.FormatInt(bucket.Key, 10)
		productCount := strconv.Itoa(bucket.DocCount)

		var shopName, shopURL string

		// Extract shop name and URL from the first hit (if available)
		if len(bucket.ShopInfo.Hits.Hits) > 0 {
			firstHit := bucket.ShopInfo.Hits.Hits[0]
			shopName = firstHit.Source.Shop.Name
			shopURL = fmt.Sprintf("https://www.pttavm.com/magaza/%s", firstHit.Source.Shop.URL)
		} else {
			shopName = "Unknown"
			shopURL = "unknown"
		}

		// Add row to CSV data
		csvData = append(csvData, []string{
			shopID,
			shopName,
			shopURL,
			productCount,
		})
	}

	// Write CSV file
	helpers.Write("assets/product/shop_aggregation_export.csv", csvData)

	fmt.Printf("✅ CSV export tamamlandı! %d mağaza verisi işlendi.\n", len(response.Aggregations.TopShops.Buckets))
	fmt.Printf("📁 Dosya konumu: assets/product/shop_aggregation_export.csv\n")

	// Print summary statistics
	totalProducts := 0
	for _, bucket := range response.Aggregations.TopShops.Buckets {
		totalProducts += bucket.DocCount
	}

	fmt.Printf("📊 İstatistikler:\n")
	fmt.Printf("   - Toplam mağaza sayısı: %d\n", len(response.Aggregations.TopShops.Buckets))
	fmt.Printf("   - Toplam ürün sayısı: %d\n", totalProducts)
	fmt.Printf("   - Diğer mağazalardaki ürün sayısı: %d\n", response.Aggregations.TopShops.SumOtherDocCount)

	if len(response.Aggregations.TopShops.Buckets) > 0 {
		topShop := response.Aggregations.TopShops.Buckets[0]
		var topShopName string
		if len(topShop.ShopInfo.Hits.Hits) > 0 {
			topShopName = topShop.ShopInfo.Hits.Hits[0].Source.Shop.Name
		} else {
			topShopName = "Unknown"
		}
		fmt.Printf("   - En fazla ürüne sahip mağaza: %s (%d ürün)\n", topShopName, topShop.DocCount)
	}
}
