package main

import (
	"flag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"my_toolbox/config"
	"my_toolbox/jobs"
	"my_toolbox/library/elasticsearch_client"
	"my_toolbox/library/log"
	"my_toolbox/library/mongodb_client"
	"my_toolbox/library/mysql_client"
	"my_toolbox/library/rabbitmq_client"
	"os"
	"reflect"
	"strings"
)

var jobName string

func main() {
	flag.StringVar(&jobName, "j", "", "Job Name")
	flag.Parse()

	if jobName == "" {
		log.GetLogger().Info("job name is required")
		os.Exit(1)
	}

	// Extra parametreleri parse et
	params := make(map[string]string)
	showHelp := false

	for _, arg := range flag.Args() {
		if arg == "help" || arg == "--help" || arg == "-h" {
			showHelp = true
			continue
		}

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				params[parts[0]] = parts[1]
			}
		}
	}

	// Config'i oku ve library'leri başlat
	cfg := config.GetConfig()

	// Library'leri config ile başlat
	mysql_client.InitWithConfig(cfg.PttavmMySQL)
	mongodb_client.InitWithConfig(cfg.PttavmMongo, cfg.ReviewMongo)
	elasticsearch_client.InitWithConfig(cfg.PttavmElasticsearch, cfg.CommissionElasticsearch)
	rabbitmq_client.InitWithConfig(cfg.PttavmRabbitMQ)

	// Mevcut library'lerden connection'ları al
	db := mysql_client.GetPttavmDB()
	mongo := mongodb_client.Get()
	reviewMongo := mongodb_client.GetReviewMongo()
	elastic := elasticsearch_client.GetElasticSearch()
	commissionElastic := elasticsearch_client.GetCommissionElasticSearch()
	rabbitMq, err := rabbitmq_client.NewRabbitMQClient()
	if err != nil {
		panic(err)
	}

	// Job instance'ı oluştur ve connection'ları set et
	jobInstance := &jobs.Job{
		DB:                db,
		Mongo:             mongo,
		ReviewMongo:       reviewMongo,
		Elastic:           elastic,
		CommissionElastic: commissionElastic,
		PttAvmRabbitMQ:    rabbitMq,
		Args:              params,
	}

	queueNameSplit := strings.Split(jobName, "_")
	job := ""
	for _, word := range queueNameSplit {
		job += cases.Title(language.English).String(cases.Lower(language.English).String(word))
	}

	ref := reflect.ValueOf(jobInstance).MethodByName(job).Call(nil)

	interf := ref[0].Interface()
	j := interf.(jobs.IJob)

	j.BindParams(j)

	// Help isteniyorsa help göster ve çık
	if showHelp {
		// Job instance'ı üzerinden ShowJobHelp çağır
		jobInstance.ShowJobHelp(j)
		return
	}

	j.Run()
}
