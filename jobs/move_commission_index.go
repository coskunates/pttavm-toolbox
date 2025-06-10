package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"io"
	"log"
	"my_toolbox/entities"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Bulk response için struct tanımı
type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

var sourceIndex = "commissions_04092023"
var targetIndex = fmt.Sprintf("commissions_%s", time.Now().Format("20060102"))

type MoveCommissionIndex struct {
	Job
}

func (mpf *MoveCommissionIndex) Run() {
	// İlk kez çalıştırıldığında aç sadece.
	mpf.moveIndex()

	mpf.handleCategoryCommissions()
	mpf.moveProductCommissions()
	mpf.moveShopCommissions()

}

func (mpf *MoveCommissionIndex) moveIndex() {
	ctx := context.Background()

	// Önce hedef sunucuda index'in var olup olmadığını kontrol et
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{targetIndex},
	}

	existsRes, err := existsReq.Do(ctx, mpf.CommissionElastic)
	if err != nil {
		log.Fatalf("Index varlığı kontrol edilemedi: %s", err)
	}
	defer existsRes.Body.Close()

	// Eğer index zaten varsa, işlemi sonlandır
	if existsRes.StatusCode == 200 {
		fmt.Printf("⚠️ '%s' index'i zaten mevcut, işlem atlanıyor.\n", targetIndex)
		return
	}

	// 1. Get mapping using Olivere Elasticsearch client
	mapping, err := mpf.Elastic.GetMapping().Index(sourceIndex).Do(ctx)
	if err != nil {
		log.Fatalf("Mapping alınamadı: %s", err)
	}

	// Extract the mapping for the source index
	indexMapping, exists := mapping[sourceIndex].(map[string]interface{})
	if !exists {
		log.Fatalf("Source index mapping not found")
	}

	// Get the 'mappings' part only
	mappingsData, exists := indexMapping["mappings"]
	if !exists {
		log.Fatalf("Mappings section not found")
	}

	// 2. Prepare the create index request body with the mapping
	createBody := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   3,
			"number_of_replicas": 1,
		},
		"mappings": mappingsData,
	}

	bodyBytes, err := json.Marshal(createBody)
	if err != nil {
		log.Fatalf("Mapping JSON marshal hatası: %s", err)
	}

	// Log the request to verify it's correct
	fmt.Printf("Creating new index '%s' with mapping\n", targetIndex)

	// 3. Create the new index with the mapping using the official Elasticsearch client
	// Use transport.Perform directly to have more control over the request
	req := esapi.IndicesCreateRequest{
		Index: targetIndex,
		Body:  bytes.NewReader(bodyBytes),
	}

	// Ensure we're using the CommissionElastic client which should have proper TLS config
	res, err := req.Do(ctx, mpf.CommissionElastic)
	if err != nil {
		log.Fatalf("Yeni index oluşturulamadı: %s", err)
	}
	defer res.Body.Close()

	// Check response
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the error response for debugging
			log.Fatalf("Error creating index: [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	fmt.Println("✅ Mapping başarıyla taşındı.")
}

func (mpf *MoveCommissionIndex) getCategories() []int {
	var cats []entities.EvoCategories
	mpf.DB.Find(&cats)

	var catIds []int
	for _, category := range cats {
		catIds = append(catIds, category.ID)
	}

	return catIds
}

func (mpf *MoveCommissionIndex) handleCategoryCommissions() {
	catIds := mpf.getCategories()

	chunks := mpf.getChunks(catIds, 20)

	for _, chunk := range chunks {
		wg := sync.WaitGroup{}
		for _, catId := range chunk {
			wg.Add(1)
			go mpf.moveCategoryCommissions(&wg, catId)
		}
		wg.Wait()
	}

}

func (mpf *MoveCommissionIndex) getChunks(records []int, cs int) [][]int {
	var divided [][]int

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}

// moveEntityData ortak fonksiyon, belirtilen entity type için verileri taşır
func (mpf *MoveCommissionIndex) moveEntityData(entityType string, entityId int) {
	ctx := context.Background()
	scrollSize := 2000
	batchSize := 100 // Bir batchteki belge sayısı

	var (
		buf        bytes.Buffer
		blk        bulkResponse
		raw        map[string]interface{}
		numErrors  int
		numIndexed int
		numItems   int
		totalItems int
	)

	// Source Elasticsearch'te scroll başlat (Olivere client ile)
	var searchResult *elastic.SearchResult

	search := mpf.Elastic.Scroll(sourceIndex).Size(scrollSize)

	// Query oluştur
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("entity_type", entityType))

	// Eğer entityId belirtilmişse (categoryId için), o filtreyi de ekle
	if entityId > 0 {
		q.Must(elastic.NewTermQuery("entity_id", entityId))
	}

	searchResult, err := search.Query(q).Pretty(false).Do(ctx)

	if err != nil {
		log.Printf("Scroll başlatılamadı (%s, id:%d): %s", entityType, entityId, err)
		return
	}

	// İlk scroll ID'yi kaydet
	scrollID := searchResult.ScrollId

	// İşlem başlangıç zamanı
	start := time.Now().UTC()

	// Batch sayaçları
	batchNum := 1
	estimatedBatches := 0
	if searchResult.TotalHits() > 0 {
		estimatedBatches = int(searchResult.TotalHits() / int64(batchSize))
		if searchResult.TotalHits()%int64(batchSize) > 0 {
			estimatedBatches++
		}
	}

	entityLabel := entityType
	if entityId > 0 {
		entityLabel = fmt.Sprintf("%s %d", entityType, entityId)
	}

	fmt.Printf("→ %s için toplam %d kayıt bulundu\n", entityLabel, searchResult.TotalHits())
	fmt.Printf("→ Batch gönderme: ")

	for {
		// Eğer sonuç yoksa döngüden çık
		if len(searchResult.Hits.Hits) == 0 {
			break
		}

		// Buffer'ı sıfırla
		buf.Reset()
		numItems = 0

		// Her bir hit için bulk index metadatası ve verisi hazırla
		for _, hit := range searchResult.Hits.Hits {
			numItems++
			totalItems++

			// Document ID oluştur
			docID := hit.Id

			// Metadata oluştur
			meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`, targetIndex, docID)
			if _, err := buf.WriteString(meta + "\n"); err != nil {
				log.Printf("Meta yazma hatası: %s", err)
				continue
			}

			// Source'u JSON olarak marshal et
			source := make(map[string]interface{})
			if err := json.Unmarshal(hit.Source, &source); err != nil {
				log.Printf("Source unmarshal hatası: %s", err)
				continue
			}

			// Veriyi JSON olarak ekle
			if err := json.NewEncoder(&buf).Encode(source); err != nil {
				log.Printf("Source encode hatası: %s", err)
				continue
			}

			// Belirlenen batch boyutuna ulaşıldığında gönder
			if numItems == batchSize || totalItems == int(searchResult.TotalHits()) {
				// Bulk request oluştur ve gönder
				res, err := mpf.CommissionElastic.Bulk(
					bytes.NewReader(buf.Bytes()),
					mpf.CommissionElastic.Bulk.WithIndex(targetIndex),
					mpf.CommissionElastic.Bulk.WithContext(ctx),
				)

				if err != nil {
					log.Printf("Batch %d gönderiminde hata: %s", batchNum, err)
					break
				}

				// Response body'yi kapat
				resBody, err := io.ReadAll(res.Body)
				res.Body.Close()

				// Tüm istek başarısız olursa
				if res.IsError() {
					numErrors += numItems
					if err := json.Unmarshal(resBody, &raw); err != nil {
						log.Printf("Response decode hatası: %s", err)
					} else {
						log.Printf(" Hata: [%d] %s: %s",
							res.StatusCode,
							raw["error"].(map[string]interface{})["type"],
							raw["error"].(map[string]interface{})["reason"],
						)
					}
				} else {
					// Başarılı yanıtı decode et ve sonuçları kontrol et
					if err := json.Unmarshal(resBody, &blk); err != nil {
						log.Printf("Response decode hatası: %s", err)
					} else {
						// Her bir belge için durum kontrolü
						for _, d := range blk.Items {
							if d.Index.Status > 201 {
								// 201'den büyük statü kodları hata
								numErrors++
								log.Printf(" Belge hatası: [%d]: %s: %s",
									d.Index.Status,
									d.Index.Error.Type,
									d.Index.Error.Reason,
								)
							} else {
								// Başarılı
								numIndexed++
							}
						}
					}
				}

				// Counter'ları güncelle
				batchNum++
				buf.Reset()
				numItems = 0
			}
		}

		// Yeni scroll sonuçlarını al
		searchResult, err = mpf.Elastic.Scroll(sourceIndex).
			ScrollId(scrollID).
			Pretty(true).
			Do(ctx)

		if err != nil {
			if err == io.EOF && searchResult != nil {
				//enp.Elastic.ClearScroll(searchResult.ScrollId)
			} else {
				log.Printf("Scroll devam hatası (%s): %s", entityType, err)
			}
			break
		}

		// Scroll ID'yi güncelle
		scrollID = searchResult.ScrollId
	}

	// İşlem bitti, scroll'u temizle
	clearScrollReq, err := mpf.Elastic.ClearScroll().ScrollId(scrollID).Do(ctx)
	if err != nil || !clearScrollReq.Succeeded {
		log.Printf("Scroll temizleme hatası (%s): %s", entityType, err)
	}

	// Sonuçları raporla
	fmt.Print("\n")
	dur := time.Since(start)

	if numErrors > 0 {
		log.Printf(
			"%s: %d belgeden %d tanesi başarıyla, %d tanesi hatayla aktarıldı (%s içinde, %d belge/sn)",
			entityLabel,
			totalItems,
			numIndexed,
			numErrors,
			dur.Truncate(time.Millisecond),
			int64(1000.0/float64(dur/time.Millisecond)*float64(numIndexed)),
		)
	} else {
		log.Printf(
			"%s: Toplam %d belge başarıyla aktarıldı (%s içinde, %d belge/sn)",
			entityLabel,
			numIndexed,
			dur.Truncate(time.Millisecond),
			int64(1000.0/float64(dur/time.Millisecond)*float64(numIndexed)),
		)
	}
}

func (mpf *MoveCommissionIndex) moveCategoryCommissions(wg *sync.WaitGroup, categoryId int) {
	defer wg.Done()
	mpf.moveEntityData("evo_category", categoryId)
}

func (mpf *MoveCommissionIndex) moveProductCommissions() {
	mpf.moveEntityData("product", 0)
}

func (mpf *MoveCommissionIndex) moveShopCommissions() {
	mpf.moveEntityData("shop", 0)
}

// min fonksiyonu
func comMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
