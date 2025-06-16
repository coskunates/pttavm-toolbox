package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"my_toolbox/elasticsearch_entities"
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
			CategoryName:   helpers.CleanString(record[3]),
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
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/rules", bytes.NewReader(payloadBytes))
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
		resp, err := csm.SendToNewCommissionService(models.CreateRuleRequest{
			Name:           fmt.Sprintf("%s Kategori Komisyonu", commission.CategoryName),
			Description:    fmt.Sprintf("%s Kategori Komisyonu", commission.CategoryName),
			Type:           "standard",
			Metadata:       nil,
			TermInDays:     1,
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
			log.GetLogger().Error("request error", err)
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
		commission := elasticsearch_entities.Commission{
			Priority:   int(source["priority"].(float64)),
			Ratio:      source["ratio"].(float64),
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
	for _, commission := range commissions {

		startDate := time.Now()
		endDate := time.Now().Add(time.Hour * 24 * 365)

		if commission.StartDate != nil {
			startDate = *commission.StartDate
		}

		if commission.EndDate != nil {
			endDate = *commission.EndDate
		}
		resp, err := csm.SendToNewCommissionService(models.CreateRuleRequest{
			Name:           fmt.Sprintf("%d Magaza Komisyonu", commission.ShopId),
			Description:    fmt.Sprintf("%d Magaza Komisyonu", commission.ShopId),
			Type:           "contract",
			Metadata:       nil,
			TermInDays:     1,
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
			log.GetLogger().Error("request error", err)
		} else {
			fmt.Println(fmt.Sprintf("%d Mağaza Komisyonu oluşturuldu. Komisyon Id: %s", commission.ShopId, resp.Data.Id))
		}
	}
}
