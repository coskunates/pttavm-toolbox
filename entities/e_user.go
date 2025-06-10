package entities

import (
	"database/sql"
	"time"
)

type EUser struct {
	ActivationDate    time.Time       `gorm:"column:activation_date"`
	FbEmail           sql.NullString  `gorm:"column:fb_email"`
	FbUserid          sql.NullString  `gorm:"column:fb_userid"`
	FbUsername        sql.NullString  `gorm:"column:fb_username"`
	HgsBalance        sql.NullFloat64 `gorm:"column:hgs_balance"`
	HgsIlkYukleme     int             `gorm:"column:hgs_ilk_yukleme"`
	HgsUser           int             `gorm:"column:hgs_user"`
	LangID            sql.NullString  `gorm:"column:lang_id"`
	LostPasswordHash  sql.NullString  `gorm:"column:lost_password_hash"`
	Mailkontrol       int             `gorm:"column:mailkontrol"`
	My                int             `gorm:"column:my"`
	PartitaIva        sql.NullString  `gorm:"column:partita_iva"`
	ProdottoEdit      int             `gorm:"column:prodotto_edit"`
	PttYetki          int             `gorm:"column:ptt_yetki"`
	SiparisYetki      int             `gorm:"column:siparis_yetki"`
	UserAccesslevel   int             `gorm:"column:user_accesslevel"`
	UserAttivato      int             `gorm:"column:user_attivato"`
	UserAvatarinfo    sql.NullString  `gorm:"column:user_avatarinfo"`
	UserAzienda       sql.NullString  `gorm:"column:user_azienda"`
	UserBirthday      time.Time       `gorm:"column:user_birthday"`
	UserCodicefiscale sql.NullString  `gorm:"column:user_codicefiscale"`
	UserCognome       string          `gorm:"column:user_cognome"`
	UserEmail         string          `gorm:"column:user_email"`
	UserID            string          `gorm:"column:user_id"`
	UserNome          string          `gorm:"column:user_nome"`
	UserNumid         int64           `gorm:"column:user_numid;primary_key"`
	UserPassword      string          `gorm:"column:user_password"`
	UserTckno         sql.NullString  `gorm:"column:user_tckno"`
	UserUID           sql.NullString  `gorm:"column:user_uid"`
}

// TableName sets the insert table name for this struct type
func (e *EUser) TableName() string {
	return "e_user"
}
