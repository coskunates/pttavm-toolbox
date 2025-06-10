package models

type CategoryCommissionTable struct {
	CategoryId     int64   `json:"category_id"`
	CommissionRate float64 `json:"commission_rate"`
}

type FixCommission struct {
	CategoryId            string `json:"category_id"`
	CategoryName          string `json:"category_name"`
	DefaultCommissionRate string `json:"default_commission_rate"`
	CommissionRate        string `json:"commission_rate"`
	LastRate              string `json:"last_rate"`
	ShopId                string `json:"shop_id"`
}
