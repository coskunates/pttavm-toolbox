package config

import (
	"encoding/json"
	"os"
)

var config Configuration

type Configuration struct {
	PttavmMySQL             MySQLConfig         `json:"pttavm_mysql"`
	PttavmMongo             MongoConfig         `json:"pttavm_mongo"`
	ReviewMongo             MongoConfig         `json:"review_mongo"`
	PttavmElasticsearch     ElasticsearchConfig `json:"pttavm_elasticsearch"`
	CommissionElasticsearch ElasticsearchConfig `json:"commission_elasticsearch"`
	PttavmRabbitMQ          RabbitMQConfig      `json:"pttavm_rabbitmq"`
	FtpServer               FtpConfig           `json:"ftp_server"`
	InsiderMail             InsiderMailConfig   `json:"insider_mail"`
	NotificationEmails      NotificationConfig  `json:"notification_emails"`
}

type InsiderMailConfig struct {
	Endpoint string `json:"endpoint"`
	AuthKey  string `json:"auth_key"`
	Timeout  int    `json:"timeout"`
}

type NotificationConfig struct {
	DryRunReports []string `json:"dry_run_reports"`
}

type FtpConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"`
}

type RabbitMQConfig struct {
	Host                 string `json:"host"`
	Port                 uint16 `json:"port"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	VirtualHost          string `json:"virtual_host"`
	ReconnectionInterval int    `json:"reconnection_interval"`
	ReconnectionAttempt  uint   `json:"reconnection_attempt"`
	ConnectionName       string `json:"connection_name"`
}

type ElasticsearchConfig struct {
	Host  string `json:"host"`
	Port  uint16 `json:"port"`
	Index string `json:"index"`
	Type  string `json:"type"`
}

type MongoConfig struct {
	DSN string `json:"dsn"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DB       string `json:"db"`
}

func init() {
	raw, err := os.ReadFile("config.json")
	if err != nil {
		panic("error while reading config")
	}

	_ = json.Unmarshal(raw, &config)
}

func GetConfig() Configuration {
	return config
}
