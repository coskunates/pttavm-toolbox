package entities

import (
	"database/sql"
	"time"
)

type EShop struct {
	AccountingCode    sql.NullString `gorm:"column:accounting_code"`
	AdresBilgisi      sql.NullString `gorm:"column:adres_bilgisi"`
	BankBranch        sql.NullString `gorm:"column:bank_branch"`
	BankName          sql.NullString `gorm:"column:bank_name"`
	CanSeeDesiLock    int            `gorm:"column:canSeeDesiLock"`
	CargoChange       sql.NullInt64  `gorm:"column:cargo_change"`
	CargoFixed        int            `gorm:"column:cargo_fixed"`
	CargoLock         int            `gorm:"column:cargo_lock"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	CreatedBy         string         `gorm:"column:created_by"`
	FgOnSale          int            `gorm:"column:fg_on_sale"`
	Gsm               string         `gorm:"column:gsm"`
	HTMLDetail        int            `gorm:"column:html_detail"`
	IbanNo            sql.NullString `gorm:"column:iban_no"`
	IsVerified        sql.NullInt64  `gorm:"column:isVerified"`
	KoKapali          int            `gorm:"column:ko_kapali"`
	Mikrokredi        int            `gorm:"column:mikrokredi"`
	NoShipping        sql.NullInt64  `gorm:"column:no_shipping"`
	PreApproval       sql.NullInt64  `gorm:"column:pre_approval"`
	PriceCompare      int            `gorm:"column:price_compare"`
	ShipmentFixed     int            `gorm:"column:shipment_fixed"`
	Shop3dattivo      int            `gorm:"column:shop_3dattivo"`
	ShopAttivo        int            `gorm:"column:shop_attivo"`
	ShopContractdate  time.Time      `gorm:"column:shop_contractdate"`
	ShopDisabled      int            `gorm:"column:shop_disabled"`
	ShopEmail         string         `gorm:"column:shop_email"`
	ShopFeriea        time.Time      `gorm:"column:shop_feriea"`
	ShopFerieda       time.Time      `gorm:"column:shop_ferieda"`
	ShopID            int            `gorm:"column:shop_id;primary_key"`
	ShopKargoLimit    sql.NullInt64  `gorm:"column:shop_kargo_limit"`
	ShopKep           sql.NullString `gorm:"column:shop_kep"`
	ShopMargine       float32        `gorm:"column:shop_margine"`
	ShopMaxprodotti   int            `gorm:"column:shop_maxprodotti"`
	ShopMaxprodotti3d int            `gorm:"column:shop_maxprodotti3d"`
	ShopMersis        sql.NullString `gorm:"column:shop_mersis"`
	ShopNocod         int            `gorm:"column:shop_nocod"`
	ShopNome          string         `gorm:"column:shop_nome"`
	ShopNote          string         `gorm:"column:shop_note"`
	ShopPersonalurl   string         `gorm:"column:shop_personalurl"`
	ShopPosizione     string         `gorm:"column:shop_posizione"`
	ShopProdottihp    int            `gorm:"column:shop_prodottihp"`
	ShopRealname      sql.NullString `gorm:"column:shop_realname"`
	ShopSellgold      sql.NullInt64  `gorm:"column:shop_sellgold"`
	ShopURL           string         `gorm:"column:shop_url"`
	ShopVendita       int            `gorm:"column:shop_vendita"`
	ShopVml           string         `gorm:"column:shop_vml"`
	SingleBox         int            `gorm:"column:single_box"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
	UpdatedBy         sql.NullString `gorm:"column:updated_by"`
	Valor             sql.NullInt64  `gorm:"column:valor"`
	WaAlert           int            `gorm:"column:waAlert"`
	XMLEditorlight    string         `gorm:"column:xml_editorlight"`
}

// TableName sets the insert table name for this struct type
func (e *EShop) TableName() string {
	return "e_shop"
}
