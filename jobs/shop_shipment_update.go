package jobs

import (
	"fmt"
	"gorm.io/gorm"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"strconv"
	"strings"
	"time"
)

var categoryIds = []string{"15", "3964", "3978", "3998", "4029", "4037", "4041", "4042", "4051", "4059", "4065", "4074", "4086", "4089", "3965", "3966", "3967", "3968", "3969",
	"3970", "3971", "3972", "3973", "3974", "3975", "3976", "3977", "3979", "3999", "4002", "4006", "4012", "4015", "4016", "4019", "4020", "4024",
	"4025", "4026", "4027", "4028", "4030", "4031", "4032", "4033", "4034", "4035", "4036", "4038", "4039", "4040", "4043", "4044", "4045", "4046",
	"4047", "4048", "4049", "4050", "4052", "4053", "4054", "4055", "4056", "4057", "4058", "4060", "4061", "4062", "4063", "4064", "4066", "4067",
	"4068", "4069", "4070", "4071", "4072", "4073", "4075", "4076", "4077", "4078", "4079", "4080", "4081", "4082", "4083", "4084", "4085", "4087",
	"4088", "4090", "4091", "4092", "4093", "3980", "3981", "3986", "3991", "3996", "4000", "4001", "4003", "4004", "4005", "4007", "4008", "4009",
	"4010", "4011", "4013", "4014", "4017", "4018", "4021", "4022", "4023", "3982", "3983", "3984", "3985", "3987", "3988", "3989", "3990", "3992",
	"3993", "3994", "3995", "3997"}

type ShopShipmentUpdate struct {
	Job
}

func (ssu *ShopShipmentUpdate) Run() {
	records, err := helpers.ReadFromCSV("assets/shop_ids.csv")
	if err != nil {
		panic(err)
	}

	for _, record := range records {
		shopId, _ := strconv.Atoi(record[0])
		ssu.updateNotInSpecificCategoryProducts(shopId)
		ssu.updateSpecificCategoryProducts(shopId)
	}
	fmt.Println(records)
}

func (ssu *ShopShipmentUpdate) updateNotInSpecificCategoryProducts(id int) {
	limit := 1000

	for {
		products := ssu.getProducts(id, limit, false)
		if len(products) <= 0 {
			break
		}

		for _, product := range products {
			ssu.updateProduct(product)
		}

		fmt.Println(fmt.Sprintf("updated product count: %d", len(products)))
	}
}

func (ssu *ShopShipmentUpdate) updateSpecificCategoryProducts(id int) {
	limit := 1000

	for {
		products := ssu.getProducts(id, limit, true)
		if len(products) <= 0 {
			break
		}

		for _, product := range products {
			ssu.updateProduct(product)
		}

		fmt.Println(fmt.Sprintf("updated product count: %d", len(products)))
	}
}

func (ssu *ShopShipmentUpdate) getProducts(shopId, limit int, isSpecific bool) []entities.EProdotto {
	priceCondition := 250
	notInCategories := "AND ep.evo_category_id NOT IN (" + strings.Join(categoryIds, ",") + ")"
	if isSpecific {
		priceCondition = 750
		notInCategories = " AND ep.evo_category_id IN (" + strings.Join(categoryIds, ",") + ")"
	}

	sql := fmt.Sprintf("SELECT ep.prodotto_id, "+
		"ep.prodotto_prezzo, "+
		"ep.prodotto_iva, "+
		"ep.prodotto_sconto, "+
		"IF(prodotto_sconto > 0, (ep.prodotto_prezzo * (1 + (ep.prodotto_iva / 100))) * (1 - (ep.prodotto_sconto / 100)), (ep.prodotto_prezzo * (1 + (ep.prodotto_iva / 100)))) as price, "+
		"ep.prodotto_attivo, "+
		"ep.shipmentfees_fixed "+
		"FROM "+
		"e_prodotto ep "+
		"WHERE "+
		"ep.shipmentfees_fixed = %d "+
		"AND ep.shop_id = %d "+
		notInCategories+
		"AND IF(prodotto_sconto > 0, (ep.prodotto_prezzo * (1 + (ep.prodotto_iva / 100))) * (1 - (ep.prodotto_sconto / 100)), (ep.prodotto_prezzo * (1 + (ep.prodotto_iva / 100)))) <= %d "+
		"ORDER BY prodotto_id LIMIT %d;", 1, shopId, priceCondition, limit)

	var products []entities.EProdotto

	ssu.DB.Raw(sql).Scan(&products)

	return products
}

func (ssu *ShopShipmentUpdate) updateProduct(product entities.EProdotto) {
	sql := ssu.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&entities.EProdotto{}).Where("prodotto_id", product.ProdottoID).Updates(map[string]interface{}{
			"last_mod":           time.Now(),
			"shipmentfees_fixed": 0,
		})
	})

	fmt.Println("SQL:", sql)

	ssu.DB.Model(&entities.EProdotto{}).Where("prodotto_id", product.ProdottoID).Updates(map[string]interface{}{
		"last_mod":           time.Now(),
		"shipmentfees_fixed": 0,
	})
}
