package jobs

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"my_toolbox/elasticsearch_entities"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"my_toolbox/mongo_entities"
	"reflect"
	"sync"
)

var (
	ExportReviewChan chan [][]string
)

var shops map[int]entities.EShop
var categories map[int]entities.EvoCategories = make(map[int]entities.EvoCategories)

type ExportReviews struct {
	Job
}

func (er *ExportReviews) Run() {
	ExportReviewChan = make(chan [][]string)

	go er.listenChannel()

	er.getCategories()

	productIds := er.getProductIds()

	chunks := er.getChunks(productIds, 3)

	for cIdx, chunk := range chunks {
		wg := sync.WaitGroup{}

		for _, productId := range chunk {
			wg.Add(1)
			go er.ProcessReviews(productId, &wg)
		}

		wg.Wait()

		fmt.Println(cIdx, "tamamlandı")
	}
}

func (er *ExportReviews) getProductIds() []int64 {
	// Veritabanı ve koleksiyon seç
	collection := er.ReviewMongo.Database("epttavm").Collection("comments")

	// Distinct sorgusu (örnek: product_id)
	values, err := collection.Distinct(context.Background(), "product_id", bson.M{"publish": 1, "product_id": bson.M{"$gt": 204454665}})
	if err != nil {
		log.GetLogger().Error("distinct product_id", err)
	}

	var productIds []int64
	// Sonuçları yazdır
	for _, v := range values {
		if v.(int32) > 204454665 {
			productIds = append(productIds, int64(v.(int32)))
		}
	}

	return productIds
}

func (er *ExportReviews) getComments(productId int64, minId int) []mongo_entities.Comment {
	filters := bson.M{"_id": bson.M{"$gt": minId}, "product_id": productId, "publish": 1}

	var comments []mongo_entities.Comment

	// Sıralama kriteri: _id alanına göre artan (1) ya da azalan (-1)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})

	cursor, err := er.ReviewMongo.Database("epttavm").Collection("comments").Find(context.Background(), filters, findOptions)
	if err != nil {
		log.GetLogger().Error(err.Error(), err)
	}

	defer cursor.Close(context.Background())

	errorDecode := cursor.All(context.Background(), &comments)
	if errorDecode != nil {
		log.GetLogger().Error(errorDecode.Error(), errorDecode)
	}

	return comments
}

func (er *ExportReviews) getCategories() {
	var evoCategories []entities.EvoCategories
	er.DB.Find(&evoCategories)

	categoryByIds := make(map[int]entities.EvoCategories)
	for _, category := range evoCategories {
		categoryByIds[category.ID] = category
	}

	categories = categoryByIds
}

func (er *ExportReviews) getShop(shopId int) entities.EShop {
	var shop entities.EShop

	if _, ok := shops[shopId]; !ok {
		er.DB.Where("shop_id", shopId).First(&shop)
	} else {
		shop = shops[shopId]
	}

	return shop
}

func (er *ExportReviews) getProduct(productId int64) *elasticsearch_entities.Product {
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("id", productId))

	ctx := context.Background()

	lim := 1000
	search := er.Elastic.Scroll("epa").Size(lim)

	var searchResult *elastic.SearchResult
	var err error

	searchResult, err = search.Query(q).Pretty(false).Do(ctx)

	if err != nil {
		if err == io.EOF && searchResult != nil {
			er.Elastic.ClearScroll(searchResult.ScrollId)
		} else {
			log.GetLogger().Error("elasticsearch error", err)
		}

		return nil
	} else {
		var product elasticsearch_entities.Product

		for _, item := range searchResult.Each(reflect.TypeOf(product)) {
			if p, ok := item.(elasticsearch_entities.Product); ok {
				product = p
				break
			}
		}

		return &product
	}
}

func (er *ExportReviews) getBrand(product *elasticsearch_entities.Product) string {
	var brand string

	if product.EvoAttributes != nil {
		for _, attribute := range product.EvoAttributes {
			// marka demek
			if attribute.AttributeType == "property" && attribute.Id == 6564 {
				brand = attribute.Values.Name
			}
		}
	}

	return brand
}

func (er *ExportReviews) getChunks(records []int64, cs int) [][]int64 {
	var divided [][]int64

	for i := 0; i < len(records); i += cs {
		end := i + cs

		if end > len(records) {
			end = len(records)
		}

		divided = append(divided, records[i:end])
	}

	return divided
}

func (er *ExportReviews) ProcessReviews(productId int64, s *sync.WaitGroup) {
	defer s.Done()

	product := er.getProduct(productId)
	if product == nil {
		return
	}
	var minId = 0

	brand := er.getBrand(product)

	for {
		comments := er.getComments(productId, minId)

		if len(comments) <= 0 {
			break
		}

		var data [][]string
		for _, comment := range comments {
			if comment.Id > minId {
				minId = comment.Id
			}

			category := categories[comment.CategoryId]
			shop := er.getShop(comment.ShopId)

			if shop.ShopID > 0 {
				data = append(data, []string{
					fmt.Sprintf("%d", comment.ProductId),
					product.Name,
					brand,
					shop.ShopNome,
					category.Name,
					comment.Comment,
					comment.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}

		ExportReviewChan <- data
	}

	fmt.Println("ProductId", productId)

	return
}

func (er *ExportReviews) listenChannel() {
	go func() {
		for {
			select {
			case reviews := <-ExportReviewChan:
				helpers.Write("assets/product/comments.csv", reviews)
			}
		}
	}()
}
