package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/text/unicode/norm"
	"io"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/models"
	"net/http"
	"strconv"
	"time"
)

type CommissionServiceMigrate struct {
	Job
}

func (csm *CommissionServiceMigrate) Run() {
	categoryCommissions := csm.getCategoryCommissions()

	csm.CreateCategoryCommissions(categoryCommissions)

	shopCommissions := csm.getShopCommissions()
	csm.CreateShopCommissions(shopCommissions)

	csm.getAndCreateProductCommissions()
}

func (csm *CommissionServiceMigrate) getCategoryCommissions() []models.CategoryCommissionTable {
	records, err := helpers.ReadFromCSV("assets/commission/category_commissions.csv")
	if err != nil {
		panic(err)
	}

	var categoryCommissionTable []models.CategoryCommissionTable
	for _, record := range records {
		categoryId, _ := strconv.ParseInt(record[2], 10, 64)
		commissionRate, _ := strconv.ParseFloat(record[4], 64)
		categoryCommissionTable = append(categoryCommissionTable, models.CategoryCommissionTable{
			CategoryId:     categoryId,
			CategoryName:   record[3],
			CommissionRate: commissionRate,
		})
	}

	return categoryCommissionTable
}

func (csm *CommissionServiceMigrate) SendToNewCommissionService(request models.CreateRuleRequest, createdBy string) (*models.RuleResponse, error) {
	// JSON'a çevir
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return nil, errors.New("JSON marshal error")
	}

	// HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// POST isteği oluştur
	req, err := http.NewRequest("POST", "https://pttavm-commission-service.pttavm.com/api/v1/rules", bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Printf("Request creation error: %v\n", err)
		return nil, errors.New("request creation error")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", createdBy)

	// İsteği gönder
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request error: %v\n", err)
		return nil, errors.New("HTTP request error")
	}
	defer resp.Body.Close()

	// HTTP status kontrolü
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Request failed with status %d: %s\n", resp.StatusCode, string(body))
		return nil, errors.New("request failed with status")
	}

	// Başarılı yanıt
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Response read error: %v\n", err)
		return nil, errors.New("response read error")
	}

	var CommissionServiceResponse models.RuleResponse
	_ = json.Unmarshal(body, &CommissionServiceResponse)

	return &CommissionServiceResponse, nil
}

func (csm *CommissionServiceMigrate) CreateCategoryCommissions(commissions []models.CategoryCommissionTable) {
	for _, commission := range commissions {
		commission.CategoryName = norm.NFC.String(commission.CategoryName)
		resp, err := csm.SendToNewCommissionService(models.CreateRuleRequest{
			Name:           fmt.Sprintf("%d - %s Kategori Komisyonu", commission.CategoryId, commission.CategoryName),
			Description:    fmt.Sprintf("%d - %s Kategori Komisyonu", commission.CategoryId, commission.CategoryName),
			Type:           "standard",
			Metadata:       nil,
			MerchantID:     nil,
			CommissionRate: commission.CommissionRate,
			RuleType:       "category",
			Filters: []models.RuleFilter{{
				Field:  "category",
				Values: []string{fmt.Sprintf("%d", commission.CategoryId)},
			}},
			Validity: models.RuleValidity{Period: models.RuleValidityPeriod{
				StartDate: time.Now(),
				EndDate:   nil,
			}},
			IsActive: true,
			Status:   "active",
		}, "system")

		if err != nil {
			log.GetLogger().Error("request error", err, zap.String("name", commission.CategoryName))
		} else {
			fmt.Println(fmt.Sprintf("%s Kategori Komisyonu oluşturuldu. Komisyon Id: %s", commission.CategoryName, resp.Data.Id))
		}
	}
}

func (csm *CommissionServiceMigrate) getShopCommissions() []elasticsearch_entities.Commission {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"entity_type": "shop",
						},
					},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.GetLogger().Error("query marshal error", err)
		return nil
	}

	res, err := csm.CommissionElastic.Search(
		csm.CommissionElastic.Search.WithContext(context.Background()),
		csm.CommissionElastic.Search.WithIndex("commissions"),
		csm.CommissionElastic.Search.WithBody(bytes.NewReader(queryJSON)),
		csm.CommissionElastic.Search.WithSize(10000),
	)
	if err != nil {
		log.GetLogger().Error("elastic search error", err)
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		log.GetLogger().Info(fmt.Sprintf("elastic search error: %s", res.Status()))
		return nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.GetLogger().Error("response decode error", err)
		return nil
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	var commissions []elasticsearch_entities.Commission

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})

		ratio := 0.0
		if source["ratio"] != nil {
			ratio = source["ratio"].(float64)
		}

		commission := elasticsearch_entities.Commission{
			Priority:   int(source["priority"].(float64)),
			Ratio:      ratio,
			EntityId:   int(source["entity_id"].(float64)),
			EntityType: source["entity_type"].(string),
			ShopId:     int(source["shop_id"].(float64)),
			CreatedBy:  source["created_by"].(string),
		}

		if startDate, ok := source["start_date"].(string); ok && startDate != "" {
			t, err := time.Parse(time.RFC3339, startDate)
			if err == nil {
				commission.StartDate = &t
			}
		}

		if endDate, ok := source["end_date"].(string); ok && endDate != "" {
			t, err := time.Parse(time.RFC3339, endDate)
			if err == nil {
				commission.EndDate = &t
			}
		}

		if ratioDate, ok := source["ratio_date"].(float64); ok && ratioDate > 0 {
			commission.RatioDate = ratioDate
		}

		if createdAt, ok := source["created_at"].(string); ok && createdAt != "" {
			t, err := time.Parse(time.RFC3339, createdAt)
			if err == nil {
				commission.CreatedAt = &t
			}
		}

		if updatedAt, ok := source["updated_at"].(string); ok && updatedAt != "" {
			t, err := time.Parse(time.RFC3339, updatedAt)
			if err == nil {
				commission.UpdatedAt = &t
			}
		}

		if updatedBy, ok := source["updated_by"].(string); ok {
			commission.UpdatedBy = updatedBy
		}

		commissions = append(commissions, commission)
	}

	return commissions
}

func (csm *CommissionServiceMigrate) CreateShopCommissions(commissions []elasticsearch_entities.Commission) {
	var shopIds []int
	for _, commission := range commissions {
		shopIds = append(shopIds, commission.ShopId)
	}

	shopInfos := csm.getShopInfos(shopIds)
	for _, commission := range commissions {

		startDate := time.Now()
		endDate := time.Now().Add(time.Hour * 24 * 365)

		if commission.StartDate != nil {
			startDate = *commission.StartDate
		}

		if commission.EndDate != nil {
			endDate = *commission.EndDate
		}

		ratio := commission.RatioDate
		if ratio == 0 && commission.Ratio > 0 {
			ratio = commission.Ratio
		}

		name := fmt.Sprintf("%d Magaza Özel Komisyonu", commission.ShopId)
		description := fmt.Sprintf("%d Magaza Özel Komisyonu", commission.ShopId)
		if _, ok := shopInfos[commission.ShopId]; ok {
			name = fmt.Sprintf("%s + Magaza Özel Komisyonu", shopInfos[commission.ShopId].ShopNome)
			description = fmt.Sprintf("%d - %s + Magaza Özel Komisyonu", commission.ShopId, shopInfos[commission.ShopId].ShopNome)
		}
		resp, err := csm.SendToNewCommissionService(models.CreateRuleRequest{
			Name:           norm.NFC.String(name),
			Description:    norm.NFC.String(description),
			Type:           "contract",
			Metadata:       nil,
			MerchantID:     nil,
			CommissionRate: commission.Ratio,
			RuleType:       "merchant_id",
			Filters: []models.RuleFilter{{
				Field:  "merchant_id",
				Values: []string{fmt.Sprintf("%d", commission.ShopId)},
			}},
			Validity: models.RuleValidity{Period: models.RuleValidityPeriod{
				StartDate: startDate,
				EndDate:   &endDate,
			}},
			IsActive: true,
			Status:   "active",
		}, commission.CreatedBy)

		if err != nil {
			log.GetLogger().Error("request error", err, zap.String("shop_id", fmt.Sprintf("%d", commission.ShopId)))
		} else {
			fmt.Println(fmt.Sprintf("%d Mağaza Komisyonu oluşturuldu. Komisyon Id: %s", commission.ShopId, resp.Data.Id))
		}
	}
}

func (csm *CommissionServiceMigrate) getAndCreateProductCommissions() {
	// İlk scroll isteği
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"entity_type": "product",
						},
					},
					{
						"range": map[string]interface{}{
							"ratio_date": map[string]interface{}{
								"gt":  0,
								"lte": 40,
							},
						},
					},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.GetLogger().Error("query marshal error", err)
		return
	}

	// İlk scroll isteği
	res, err := csm.CommissionElastic.Search(
		csm.CommissionElastic.Search.WithContext(context.Background()),
		csm.CommissionElastic.Search.WithIndex("commissions"),
		csm.CommissionElastic.Search.WithBody(bytes.NewReader(queryJSON)),
		csm.CommissionElastic.Search.WithSize(1000),
		csm.CommissionElastic.Search.WithScroll(time.Minute*5),
	)
	if err != nil {
		log.GetLogger().Error("elastic search error", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.GetLogger().Info(fmt.Sprintf("elastic search error: %s", res.Status()))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.GetLogger().Error("response decode error", err)
		return
	}

	scrollID := result["_scroll_id"].(string)
	totalProcessed := 0

	for {
		hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

		if len(hits) == 0 {
			break
		}

		var commissions []elasticsearch_entities.Commission

		for _, hit := range hits {
			source := hit.(map[string]interface{})["_source"].(map[string]interface{})

			ratio := 0.0
			if source["ratio"] != nil {
				ratio = source["ratio"].(float64)
			}
			commission := elasticsearch_entities.Commission{
				Priority:   int(source["priority"].(float64)),
				Ratio:      ratio,
				EntityId:   int(source["entity_id"].(float64)),
				EntityType: source["entity_type"].(string),
				ShopId:     int(source["shop_id"].(float64)),
			}

			if startDate, ok := source["start_date"].(string); ok && startDate != "" {
				t, err := time.Parse(time.RFC3339, startDate)
				if err == nil {
					commission.StartDate = &t
				}
			}

			if endDate, ok := source["end_date"].(string); ok && endDate != "" {
				t, err := time.Parse(time.RFC3339, endDate)
				if err == nil {
					commission.EndDate = &t
				}
			}

			if ratioDate, ok := source["ratio_date"].(float64); ok && ratioDate > 0 {
				commission.RatioDate = ratioDate
			}

			if createdBy, ok := source["created_by"].(string); ok && createdBy != "" {
				commission.CreatedBy = createdBy
			} else {
				commission.CreatedBy = source["updated_by"].(string)
			}

			if createdAt, ok := source["created_at"].(string); ok && createdAt != "" {
				t, err := time.Parse(time.RFC3339, createdAt)
				if err == nil {
					commission.CreatedAt = &t
				}
			}

			if updatedAt, ok := source["updated_at"].(string); ok && updatedAt != "" {
				t, err := time.Parse(time.RFC3339, updatedAt)
				if err == nil {
					commission.UpdatedAt = &t
				}
			}

			if updatedBy, ok := source["updated_by"].(string); ok {
				commission.UpdatedBy = updatedBy
			}

			commissions = append(commissions, commission)
		}

		// Bu batch'i yeni servise gönder
		csm.CreateProductCommissions(commissions)

		totalProcessed += len(commissions)
		log.GetLogger().Info(fmt.Sprintf("Product commissions processed: %d", totalProcessed))

		// Sonraki scroll isteği
		scrollQuery := map[string]interface{}{
			"scroll":    "5m",
			"scroll_id": scrollID,
		}

		scrollQueryJSON, err := json.Marshal(scrollQuery)
		if err != nil {
			log.GetLogger().Error("scroll query marshal error", err)
			break
		}

		scrollRes, err := csm.CommissionElastic.Scroll(
			csm.CommissionElastic.Scroll.WithContext(context.Background()),
			csm.CommissionElastic.Scroll.WithBody(bytes.NewReader(scrollQueryJSON)),
		)
		if err != nil {
			log.GetLogger().Error("scroll request error", err)
			break
		}
		defer scrollRes.Body.Close()

		if scrollRes.IsError() {
			log.GetLogger().Info(fmt.Sprintf("scroll response error: %s", scrollRes.Status()))
			break
		}

		if err := json.NewDecoder(scrollRes.Body).Decode(&result); err != nil {
			log.GetLogger().Error("scroll response decode error", err)
			break
		}

		scrollID = result["_scroll_id"].(string)
	}

	// Scroll context'ini temizle
	clearScrollQuery := map[string]interface{}{
		"scroll_id": []string{scrollID},
	}

	clearScrollJSON, err := json.Marshal(clearScrollQuery)
	if err == nil {
		clearRes, err := csm.CommissionElastic.ClearScroll(
			csm.CommissionElastic.ClearScroll.WithContext(context.Background()),
			csm.CommissionElastic.ClearScroll.WithBody(bytes.NewReader(clearScrollJSON)),
		)
		if err == nil {
			clearRes.Body.Close()
		}
	}

	log.GetLogger().Info(fmt.Sprintf("Total product commissions processed: %d", totalProcessed))
}

func (csm *CommissionServiceMigrate) CreateProductCommissions(commissions []elasticsearch_entities.Commission) {
	var shopIds []int
	for _, commission := range commissions {
		shopIds = append(shopIds, commission.ShopId)
	}

	shopInfos := csm.getShopInfos(shopIds)

	for _, commission := range commissions {
		startDate := time.Now()
		endDate := time.Now().Add(time.Hour * 24 * 365)

		if commission.StartDate != nil {
			startDate = *commission.StartDate
		}

		if commission.EndDate != nil {
			endDate = *commission.EndDate
		}

		shopId := int32(commission.ShopId)

		name := fmt.Sprintf("%d Ürün Komisyonu", commission.EntityId)
		description := fmt.Sprintf("%d Ürün Komisyonu", commission.EntityId)
		if _, ok := shopInfos[commission.ShopId]; ok {
			name = fmt.Sprintf("%s - %d + Ürün Özel Komisyonu", shopInfos[commission.ShopId].ShopNome, commission.EntityId)
			description = fmt.Sprintf("%s - %d + Ürün Özel Komisyonu", shopInfos[commission.ShopId].ShopNome, commission.EntityId)
		}

		resp, err := csm.SendToNewCommissionService(models.CreateRuleRequest{
			Name:           norm.NFC.String(name),
			Description:    norm.NFC.String(description),
			Type:           "promotion",
			Metadata:       nil,
			MerchantID:     &shopId,
			CommissionRate: commission.RatioDate,
			RuleType:       "product",
			Filters: []models.RuleFilter{{
				Field:  "product",
				Values: []string{fmt.Sprintf("%d", commission.EntityId)},
			}},
			Validity: models.RuleValidity{Period: models.RuleValidityPeriod{
				StartDate: startDate,
				EndDate:   &endDate,
			}},
			IsActive: true,
			Status:   "active",
		}, commission.CreatedBy)

		if err != nil {
			log.GetLogger().Error("request error", err, zap.String("product_id", fmt.Sprintf("%d", commission.EntityId)))
		} else {
			_ = resp.Data.Id
			//fmt.Println(fmt.Sprintf("%d Ürün Komisyonu oluşturuldu. Komisyon Id: %s", commission.EntityId, resp.Data.Id))
		}
	}
}

func (csm *CommissionServiceMigrate) getShopInfos(ids []int) map[int]entities.EShop {
	var shopInfos []entities.EShop
	csm.DB.
		Where("shop_id IN (?)", ids).
		Limit(len(ids)).
		Find(&shopInfos)

	shopIdMap := make(map[int]entities.EShop, len(ids))
	for _, shopInfo := range shopInfos {
		shopIdMap[shopInfo.ShopID] = shopInfo
	}

	return shopIdMap
}
