package mongodb_client

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"my_toolbox/config"
	myLog "my_toolbox/library/log"
	"time"
)

var (
	client       *mongo.Client
	reviewClient *mongo.Client
	pttavmConfig config.MongoConfig
	reviewConfig config.MongoConfig
)

// InitWithConfig config ile MongoDB client'larını başlatır
func InitWithConfig(pttavm, review config.MongoConfig) {
	pttavmConfig = pttavm
	reviewConfig = review
}

func Get() *mongo.Client {
	if client == nil {
		dsn := pttavmConfig.DSN
		// Eğer config set edilmemişse eski değerleri kullan
		if dsn == "" {
			panic("pttavmConfig.DSN is null")
		}

		opts := options.Client().ApplyURI(dsn).SetTimeout(30 * time.Second)
		mongoClient, err := mongo.Connect(context.Background(), opts)

		if err != nil {
			myLog.GetLogger().Error("MongoDB connection failed", err)
			panic(err)
		}

		client = mongoClient
		myLog.GetLogger().Info("MongoDB (default) connected: " + dsn)
	}

	return client
}

func GetReviewMongo() *mongo.Client {
	if reviewClient == nil {
		dsn := reviewConfig.DSN
		// Eğer config set edilmemişse eski değerleri kullan
		if dsn == "" {
			panic("reviewConfig.DSN is null")
		}

		opts := options.Client().ApplyURI(dsn).SetTimeout(30 * time.Second)
		mongoClient, err := mongo.Connect(context.Background(), opts)

		if err != nil {
			myLog.GetLogger().Error("Review MongoDB connection failed", err)
			panic(err)
		}

		reviewClient = mongoClient
		myLog.GetLogger().Info("MongoDB (review) connected: " + dsn)
	}

	return reviewClient
}
