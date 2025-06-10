package elasticsearch_entities

import "time"

type DiscountPartner struct {
	Type string  `json:"type"`
	Rate float64 `json:"rate"`
}

type Product struct {
	Id          int64 `json:"id"`
	EvoCategory struct {
		Id                int    `json:"id"`
		Name              string `json:"name"`
		URL               string `json:"url"`
		Breadcrumb        string `json:"breadcrumb"`
		BreadcrumbText    string `json:"breadcrumb_text"`
		BreadcrumbURL     string `json:"breadcrumb_url"`
		BreadcrumbURLText string `json:"breadcrumb_url_text"`
		BreadcrumbURLKey  string `json:"breadcrumb_url_key"`
		ParentID          int    `json:"parent_id"`
	} `json:"evo_category"`
	EvoAttributes    []Attribute `json:"evo_attributes"`
	Name             string      `json:"name"`
	URL              string      `json:"url"`
	ShopID           int         `json:"shop_id"`
	ShopName         string      `json:"shop_name"`
	ShopSlug         string      `json:"shop_slug"`
	CategoryName     string      `json:"category_name"`
	CategoryURL      string      `json:"category_url"`
	SubCategoryName  string      `json:"sub_category_name"`
	SubCategoryURL   string      `json:"sub_category_url"`
	Category         int         `json:"category"`
	SubCategory      int         `json:"sub_category"`
	Price            float64     `json:"price"`
	ShipPrice        float64     `json:"ship_price"`
	ListingPrice     float64     `json:"listing_price"`
	PriceExcludedVat float64     `json:"prodotto_prezzo"`
	TotalPrice       float64     `json:"total_price"`
	OriginalPrice    float64     `json:"original_price"`
	SellCount        int         `json:"sell_count"`
	PriorityDiscount struct {
		Rate            float64           `json:"rate"`
		FixedPrice      float64           `json:"fixed_price"`
		StartedAt       time.Time         `json:"started_at"`
		DiscountPartner []DiscountPartner `json:"discount_partner"`
		Id              string            `json:"id"`
		Type            string            `json:"type"`
		EndedAt         time.Time         `json:"ended_at"`
	} `json:"priority_discount"`
	TaxRate          float64 `json:"tax_rate"`
	ParentID         int     `json:"parent_id"`
	Rating           float64 `json:"rating"`
	RatingCount      float64 `json:"rating_count"`
	IsAdult          bool    `json:"is_adult"`
	ImagePassive     bool    `json:"image_passive"`
	ShopMarketing    bool    `json:"shop_marketing"`
	Stock            int     `json:"stock"`
	OnSale           bool    `json:"on_sale"`
	OnFgSale         bool    `json:"on_fg_sale"`
	GoogleBrand      string  `json:"google_brand"`
	Brand            string  `json:"brand"`
	Gtin             string  `json:"gtin"`
	Mpn              string  `json:"mpn"`
	Ean              string  `json:"ean"`
	ShortDescription string  `json:"short_description"`
	ShopActive       bool    `json:"shop_active"`
	Active           bool    `json:"active"`
	GoogleLabel0     string  `json:"google_label_0"`
	GoogleLabel1     string  `json:"google_label_1"`
	GoogleCatID      int     `json:"google_cat_id"`
	EvoStock         []struct {
		Combination string  `json:"combination"`
		Stock       int     `json:"stock"`
		DiffPrice   float64 `json:"differenza_prezzo"`
	} `json:"evo_stock"`
	Image []struct {
		Num170 string `json:"170"`
		Num592 string `json:"592"`
	} `json:"image"`
	Thumbnail       string `json:"thumbnail"`
	BigThumbnail    string `json:"big_thumbnail"`
	FullDescription string `json:"full_description"`
	Comments        []struct {
		Id          int       `json:"id"`
		UserID      int       `json:"user_id"`
		Rate        int       `json:"rate"`
		Comment     string    `json:"comment"`
		CommentDate time.Time `json:"comment_date"`
		FirstName   string    `json:"first_name"`
		LastName    string    `json:"last_name"`
		Username    string    `json:"username"`
		Gender      string    `json:"gender"`
		Birthday    time.Time `json:"birthday"`
	} `json:"comments"`
	Properties   []int `json:"properties"`
	FastShipping int   `json:"fast_shipping"`
	Shop         struct {
		Id          int     `json:"id"`
		Name        string  `json:"name"`
		URL         string  `json:"url"`
		Rating      float64 `json:"rating"`
		RatingCount int     `json:"rating_count"`
		MaxProduct  int     `json:"max_product"`
		IsVerified  bool    `json:"is_verified"`
		OnSale      bool    `json:"on_sale"`
		OnFgSale    bool    `json:"on_fg_sale"`
		Logo        string  `json:"logo"`
	} `json:"shop"`
	Installments []Installment     `json:"installments"`
	Installment  []InstallmentItem `json:"installment"`
	ProductType  string            `json:"product_type"`
	UpdatedAt    time.Time         `json:"__updated_at"`
}

type Attribute struct {
	AttributeType string `json:"attribute_type"`
	Id            int    `json:"id"`
	Name          string `json:"name"`
	NameText      string `json:"name_text"`
	NameKey       string `json:"name_key"`
	KeyValue      string `json:"key_value"`
	Values        struct {
		Id       int     `json:"id"`
		Name     string  `json:"name"`
		NameText string  `json:"name_text"`
		NameKey  string  `json:"name_key"`
		KeyValue string  `json:"key_value"`
		Stock    int     `json:"stock"`
		AddPrice float64 `json:"add_price"`
	} `json:"values,omitempty"`
}

type Installment struct {
	Name         string            `json:"name"`
	BankCode     int               `json:"bank_code"`
	Installments []InstallmentItem `json:"installments"`
}

type InstallmentItem struct {
	Value int     `json:"value"`
	Ratio float64 `json:"ratio"`
}
