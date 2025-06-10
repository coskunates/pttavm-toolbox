package entities

type EProdottoContent struct {
	LangID                  string `gorm:"column:lang_id;primary_key"`
	ProdottoAbstract        string `gorm:"column:prodotto_abstract"`
	ProdottoAttivo          int    `gorm:"column:prodotto_attivo"`
	ProdottoDescrizione     string `gorm:"column:prodotto_descrizione"`
	ProdottoExturl          string `gorm:"column:prodotto_exturl"`
	ProdottoFotodesc1       string `gorm:"column:prodotto_fotodesc1"`
	ProdottoFotodesc2       string `gorm:"column:prodotto_fotodesc2"`
	ProdottoFotodesc3       string `gorm:"column:prodotto_fotodesc3"`
	ProdottoFotodesc4       string `gorm:"column:prodotto_fotodesc4"`
	ProdottoFotodesc5       string `gorm:"column:prodotto_fotodesc5"`
	ProdottoHighlights      string `gorm:"column:prodotto_highlights"`
	ProdottoID              int64  `gorm:"column:prodotto_id;primary_key"`
	ProdottoMetadescription string `gorm:"column:prodotto_metadescription"`
	ProdottoNome            string `gorm:"column:prodotto_nome"`
	ProdottoURL             string `gorm:"column:prodotto_url"`
}

// TableName sets the insert table name for this struct type
func (e *EProdottoContent) TableName() string {
	return "e_prodotto_content"
}
