package elasticsearch_client

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/olivere/elastic/v7"
	"my_toolbox/config"
	"my_toolbox/library/log"
	"sync"
	"time"
)

var (
	legacyClient     *elastic.Client
	commissionClient *elasticsearch.Client
	onceLegacy       sync.Once
	onceCommission   sync.Once
	pttavmConfig     config.ElasticsearchConfig
	commissionConfig config.ElasticsearchConfig
)

// InitWithConfig config ile Elasticsearch client'larını başlatır
func InitWithConfig(pttavm, commission config.ElasticsearchConfig) {
	pttavmConfig = pttavm
	commissionConfig = commission
}

func GetElasticSearch() *elastic.Client {
	onceLegacy.Do(func() {
		// Config'den URL oluştur
		url := fmt.Sprintf("http://%s:%d", pttavmConfig.Host, pttavmConfig.Port)
		// Eğer config set edilmemişse eski değerleri kullan
		if pttavmConfig.Host == "" {
			panic("pttavmConfig.Host is required")
		}

		esClient, err := elastic.NewClient(
			elastic.SetSniff(false),
			elastic.SetURL(url),
			elastic.SetHealthcheckInterval(5*time.Second),
		)

		if err != nil {
			log.GetLogger().Error("elastic initialize error", err)
		} else {
			log.GetLogger().Info("Elasticsearch v7 connected: " + url)
		}

		legacyClient = esClient
	})

	return legacyClient
}

func GetCommissionElasticSearch() *elasticsearch.Client {
	onceCommission.Do(func() {
		// Config'den URL oluştur
		url := fmt.Sprintf("http://%s:%d", commissionConfig.Host, commissionConfig.Port)
		// Eğer config set edilmemişse eski değerleri kullan
		if commissionConfig.Host == "" {
			panic("commissionConfig.Host is required")
		}

		esClient, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{url},
		})
		if err != nil {
			log.GetLogger().Error("elastic8 initialize error", err)
		} else {
			log.GetLogger().Info("Elasticsearch v8 connected: " + url)
		}

		commissionClient = esClient
	})

	return commissionClient
}
