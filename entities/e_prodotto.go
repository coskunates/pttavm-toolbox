package entities

import (
	"database/sql"
	"time"
)

type EProdotto struct {
	AmountOfSupplierCommission sql.NullFloat64 `gorm:"column:amount_of_supplier_commission"`
	Badges                     sql.NullString  `gorm:"column:badges"`
	CatprodID                  sql.NullInt64   `gorm:"column:catprod_id"`
	Desi                       sql.NullFloat64 `gorm:"column:desi"`
	DoNotCheckLockDesi         int             `gorm:"column:do_not_check_lock_desi"`
	EvoCategoryID              int             `gorm:"column:evo_category_id"`
	FgOnSale                   int             `gorm:"column:fg_on_sale"`
	GhostCampaign              int             `gorm:"column:ghost_campaign"`
	GhostPrice                 sql.NullFloat64 `gorm:"column:ghost_price"`
	GhostSconto                sql.NullFloat64 `gorm:"column:ghost_sconto"`
	GoldCommission             sql.NullFloat64 `gorm:"column:gold_commission"`
	ImgCount                   int             `gorm:"column:img_count"`
	LastIP                     string          `gorm:"column:last_ip"`
	LastMod                    time.Time       `gorm:"column:last_mod"`
	MarcaID                    int             `gorm:"column:marca_id"`
	NoShipping                 sql.NullInt64   `gorm:"column:no_shipping"`
	ProdottoAltezza            float32         `gorm:"column:prodotto_altezza"`
	ProdottoAttivo             string          `gorm:"column:prodotto_attivo"`
	ProdottoCodice             string          `gorm:"column:prodotto_codice"`
	ProdottoEanIsbn            string          `gorm:"column:prodotto_ean_isbn"`
	ProdottoID                 int             `gorm:"column:prodotto_id;primary_key"`
	ProdottoIva                float64         `gorm:"column:prodotto_iva"`
	ProdottoMargine            int             `gorm:"column:prodotto_margine"`
	ProdottoPeso               float32         `gorm:"column:prodotto_peso"`
	ProdottoPosizione          int             `gorm:"column:prodotto_posizione"`
	ProdottoPrezzo             float64         `gorm:"column:prodotto_prezzo"`
	ProdottoQuantita           int             `gorm:"column:prodotto_quantita"`
	ProdottoSconto             float64         `gorm:"column:prodotto_sconto"`
	ProdottoScontoEsimple      int             `gorm:"column:prodotto_sconto_esimple"`
	ProdottoScontovip          int             `gorm:"column:prodotto_scontovip"`
	ProdottoScontovipEsimple   int             `gorm:"column:prodotto_scontovip_esimple"`
	ProdottoShopID             string          `gorm:"column:prodotto_shop_id"`
	ProdottoSizex              int             `gorm:"column:prodotto_sizex"`
	ProdottoSizey              int             `gorm:"column:prodotto_sizey"`
	ProdottoSizez              int             `gorm:"column:prodotto_sizez"`
	ProdottoType               int             `gorm:"column:prodotto_type"`
	ProdottoVisibility         int             `gorm:"column:prodotto_visibility"`
	SellerStockCode            string          `gorm:"column:seller_stock_code"`
	ShipmentfeesPrice          float32         `gorm:"column:shipmentfees_price"`
	ShopCatprodID              sql.NullInt64   `gorm:"column:shop_catprod_id"`
	ShopID                     int             `gorm:"column:shop_id"`
	SubcatprodID               sql.NullInt64   `gorm:"column:subcatprod_id"`
	TempID                     int             `gorm:"column:temp_id"`
	UrunBarkod                 string          `gorm:"column:urun_barkod"`
	WarrantyCompany            string          `gorm:"column:warranty_company"`
	WarrantyDuration           int             `gorm:"column:warranty_duration"`
}

// TableName sets the insert table name for this struct type
func (e *EProdotto) TableName() string {
	return "e_prodotto"
}
