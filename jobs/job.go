package jobs

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"my_toolbox/library/rabbitmq_client"
	"reflect"
	"strconv"
	"strings"
)

type IJob interface {
	BindParams(interface{})
	Run()
}

type Job struct {
	// Database connections - main.go'da set edilir
	DB                *gorm.DB
	Mongo             *mongo.Client
	ReviewMongo       *mongo.Client
	Elastic           *elastic.Client
	CommissionElastic *elasticsearch.Client
	PttAvmRabbitMQ    *rabbitmq_client.RabbitMQ

	Args map[string]string

	// Common parameters - tüm job'larda kullanılabilir
	Limit   int  `param:"limit" help:"Batch processing limit per iteration" default:"1000"`
	DryRun  bool `param:"dry_run" help:"Dry run mode - show what would be done without making changes" default:"false"`
	Verbose bool `param:"verbose" help:"Enable verbose output for detailed logging" default:"false"`
}

// ShowJobHelp job'ın parametrelerini ve açıklamalarını gösterir
func (j *Job) ShowJobHelp(target interface{}) {
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	jobName := typ.Name()

	fmt.Printf("=== %s Job Help ===\n\n", jobName)
	fmt.Println("Usage:")
	fmt.Printf("  ./toolbox -j %s [options]\n\n", convertJobNameToFlag(jobName))

	// Önce common parameters göster
	fmt.Println("Common Parameters:")
	j.showStructParams(reflect.ValueOf(j).Elem())

	// Sonra job-specific parameters göster
	fmt.Println("\nJob-Specific Parameters:")
	hasJobParams := j.showStructParams(val)

	if !hasJobParams {
		fmt.Println("  No job-specific parameters available.")
	}

	fmt.Println("\nExamples:")
	fmt.Printf("  ./toolbox -j %s dry_run=true verbose=true\n", convertJobNameToFlag(jobName))
	fmt.Printf("  ./toolbox -j %s limit=500 param1=value1\n", convertJobNameToFlag(jobName))
	fmt.Println()
}

func (j *Job) showStructParams(val reflect.Value) bool {
	typ := val.Type()
	hasParams := false

	for i := 0; i < val.NumField(); i++ {
		fieldType := typ.Field(i)

		// param tag'ini kontrol et
		paramTag := fieldType.Tag.Get("param")
		if paramTag == "" {
			continue
		}

		hasParams = true
		helpTag := fieldType.Tag.Get("help")
		defaultTag := fieldType.Tag.Get("default")

		// Parametre adını flag formatına çevir
		flagName := paramTag

		// Type bilgisi
		typeStr := getTypeString(fieldType.Type)

		fmt.Printf("  %-20s %s", flagName, typeStr)

		if helpTag != "" {
			fmt.Printf(" - %s", helpTag)
		}

		if defaultTag != "" {
			fmt.Printf(" (default: %s)", defaultTag)
		}

		fmt.Println()
	}

	return hasParams
}

// Helper fonksiyonlar
func convertJobNameToFlag(jobName string) string {
	// UpdateShopProducts -> update_shop_products
	result := ""
	for i, r := range jobName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return strings.ToLower(result)
}

func getTypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "<string>"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "<int>"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "<uint>"
	case reflect.Bool:
		return "<bool>"
	case reflect.Float32, reflect.Float64:
		return "<float>"
	default:
		return "<value>"
	}
}

// BindParams otomatik olarak Args'dan struct field'larına değer atar
func (j *Job) BindParams(target interface{}) {
	if j.Args == nil {
		return
	}

	// Önce Job struct'ındaki common field'ları set et
	j.bindStructFields(j)

	// Sonra target job'ındaki specific field'ları set et
	j.bindStructFields(target)
}

func (j *Job) bindStructFields(target interface{}) {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr {
		return // Pointer olmalı ki set edebilelim
	}

	val = val.Elem() // Pointer'dan actual value'ya geç
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// param tag'ini kontrol et
		paramTag := fieldType.Tag.Get("param")
		if paramTag == "" {
			continue
		}

		// Args'dan değeri al
		argValue, exists := j.Args[paramTag]
		if !exists {
			continue
		}

		// Field'ı set edilebilir mi kontrol et
		if !field.CanSet() {
			continue
		}

		// Type'a göre convert et ve set et
		switch field.Kind() {
		case reflect.String:
			field.SetString(argValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(argValue, 10, 64); err == nil {
				field.SetInt(intVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintVal, err := strconv.ParseUint(argValue, 10, 64); err == nil {
				field.SetUint(uintVal)
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(argValue); err == nil {
				field.SetBool(boolVal)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(argValue, 64); err == nil {
				field.SetFloat(floatVal)
			}
		}
	}
}

// Helper metodlar (backward compatibility için)
func (j *Job) GetArg(key string, defaultValue string) string {
	if j.Args == nil {
		return defaultValue
	}
	if val, exists := j.Args[key]; exists {
		return val
	}
	return defaultValue
}

func (j *Job) GetIntArg(key string, defaultValue int) int {
	if j.Args == nil {
		return defaultValue
	}
	if str, exists := j.Args[key]; exists {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

func (j *Job) GetBoolArg(key string, defaultValue bool) bool {
	if j.Args == nil {
		return defaultValue
	}
	if str, exists := j.Args[key]; exists {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return defaultValue
}

func (j *Job) CheckFbUsers() IJob {
	return &CheckFBUsers{Job: *j}
}

func (j *Job) DuplicateBrandDetector() IJob {
	return &DuplicateBrandDetector{Job: *j}
}

func (j *Job) BrandNameFixer() IJob {
	return &BrandNameFixer{Job: *j}
}

func (j *Job) PropertyUpdateSql() IJob {
	return &PropertyUpdateSql{Job: *j}
}

func (j *Job) MoveProductsFromDeletedBrands() IJob {
	return &MoveProductsFromDeletedBrands{Job: *j}
}

func (j *Job) ActiveProductUpdate() IJob {
	return &ActiveProductUpdate{Job: *j}
}

func (j *Job) PassiveProductUpdate() IJob {
	return &PassiveProductUpdate{Job: *j}
}

func (j *Job) CommissionFixer() IJob {
	return &CommissionFixer{Job: *j}
}

func (j *Job) DetectChangedCommissions() IJob {
	return &DetectChangedCommissions{Job: *j}
}

func (j *Job) ProductListenerDeletedProductCleaner() IJob {
	return &ProductListenerDeletedProductCleaner{Job: *j}
}

func (j *Job) ExportCommissionsFromCsv() IJob {
	return &ExportCommissionsFromCsv{Job: *j}
}

func (j *Job) UpdateCategoryCommissionOnDb() IJob {
	return &UpdateCategoryCommissionOnDb{Job: *j}
}

func (j *Job) ExportCategoryCommissionCheckTable() IJob {
	return &ExportCategoryCommissionCheckTable{Job: *j}
}
func (j *Job) UpdateShopProducts() IJob {
	return &UpdateShopProducts{Job: *j}
}

func (j *Job) ExportNoBrandProducts() IJob {
	return &ExportNoBrandProducts{Job: *j}
}

func (j *Job) ChunkNoBrandProducts() IJob {
	return &ChunkNoBrandProducts{Job: *j}
}

func (j *Job) ExportNoImageProducts() IJob {
	return &ExportNoImageProducts{Job: *j}
}

func (j *Job) ExportNoImageProductsShops() IJob {
	return &ExportNoImageProductsShops{Job: *j}
}

func (j *Job) ExportNoImageProductsSorted() IJob {
	return &ExportNoImageProductsSorted{Job: *j}
}

func (j *Job) ExportImageExistsProducts() IJob {
	return &ExportImageExistsProducts{Job: *j}
}

func (j *Job) ProductUpdateIndexer() IJob {
	return &ProductUpdateIndexer{Job: *j}
}

func (j *Job) ProductTaxByCategoriesUpdate() IJob {
	return &ProductTaxByCategoriesUpdate{Job: *j}
}

func (j *Job) ImageExistsWithError() IJob {
	return &ImageExistsWithError{Job: *j}
}

func (j *Job) RequeueProductImageErrors() IJob {
	return &RequeueProductImageErrors{Job: *j}
}

func (j *Job) ImageSendRequestToServer() IJob {
	return &ImageSendRequestToServer{Job: *j}
}

func (j *Job) DetectDuplicateProductImages() IJob {
	return &DetectDuplicateProductImages{Job: *j}
}

func (j *Job) ProcessDuplicateProductImages() IJob {
	return &ProcessDuplicateProductImages{Job: *j}
}

func (j *Job) ClearPriceAndStockLocks() IJob {
	return &ClearPriceAndStockLocks{Job: *j}
}

func (j *Job) ProductListenerNotExistedProductCleaner() IJob {
	return &ProductListenerNotExistedProductCleaner{Job: *j}
}

func (j *Job) UpdateProductUrls() IJob {
	return &UpdateProductUrls{Job: *j}
}

func (j *Job) CopyCommentToNewProduct() IJob {
	return &CopyCommentToNewProduct{Job: *j}
}

func (j *Job) CommissionFixForShops() IJob {
	return &CommissionFixForShops{Job: *j}
}

func (j *Job) DetectEmptyUrlImages() IJob {
	return &DetectEmptyUrlImages{Job: *j}
}

func (j *Job) ProcessEmptyUrlImages() IJob {
	return &ProcessEmptyUrlImages{Job: *j}
}

func (j *Job) AddDeletedToProducts() IJob {
	return &AddDeletedToProducts{Job: *j}
}

func (j *Job) ShopShipmentUpdate() IJob {
	return &ShopShipmentUpdate{Job: *j}
}

func (j *Job) ProductsWithoutCategory() IJob {
	return &ProductsWithoutCategory{Job: *j}
}

func (j *Job) ProductUrlFixer() IJob {
	return &ProductUrlFixer{Job: *j}
}

func (j *Job) AttributeExporter() IJob {
	return &AttributeExporter{Job: *j}
}

func (j *Job) AttributeNameFix() IJob {
	return &AttributeNameFix{Job: *j}
}

func (j *Job) AssignBarcodes() IJob {
	return &AssignBarcodes{Job: *j}
}

func (j *Job) MoveProductLocks() IJob {
	return &MoveProductLocks{Job: *j}
}

func (j *Job) UpdateNoContentProducts() IJob {
	return &UpdateNoContentProducts{Job: *j}
}

func (j *Job) ExportReviews() IJob {
	return &ExportReviews{Job: *j}
}

func (j *Job) MoveCommissionIndex() IJob {
	return &MoveCommissionIndex{Job: *j}
}

func (j *Job) ProcessImageUrlWithVersion() IJob {
	return &ProcessImageUrlWithVersion{Job: *j}
}

func (j *Job) ExportShopAggregation() IJob {
	return &ExportShopAggregation{Job: *j}
}

func (j *Job) ProductListenerCleaner() IJob {
	return &ProductListenerCleaner{Job: *j}
}

func (j *Job) CommissionServiceMigrate() IJob {
	return &CommissionServiceMigrate{Job: *j}
}
