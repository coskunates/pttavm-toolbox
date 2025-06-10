package mysql_client

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"my_toolbox/config"
	"my_toolbox/library/log"
)

var (
	client       *gorm.DB
	clientPttavm *gorm.DB
	mysqlConfig  config.MySQLConfig
)

// InitWithConfig config ile MySQL client'ı başlatır
func InitWithConfig(cfg config.MySQLConfig) {
	mysqlConfig = cfg
}

func GetPttavmDB() *gorm.DB {
	if clientPttavm == nil {
		// Config'den bilgileri al, yoksa default'ları kullan
		user := mysqlConfig.Username
		password := mysqlConfig.Password
		host := mysqlConfig.Host
		port := mysqlConfig.Port
		db := mysqlConfig.DB

		// Eğer config set edilmemişse eski değerleri kullan
		if user == "" {
			panic("mysqlConfig is empty")
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, db)
		dataBase, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		if err != nil {
			log.GetLogger().Error("gorm mysql new client error", err)
		}

		clientPttavm = dataBase
		log.GetLogger().Info(fmt.Sprintf("MySQL connected to %s:%d/%s", host, port, db))
	}

	return clientPttavm
}
