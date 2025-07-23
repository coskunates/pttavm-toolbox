package jobs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"my_toolbox/config"
	"my_toolbox/helpers"
	"my_toolbox/library/ftp_client"
	"my_toolbox/library/log"
	"my_toolbox/library/mail_client"
	"os"
	"path/filepath"
	"time"
)

type DryRunLog struct {
	Id                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderId               string             `json:"order_id" bson:"order_id"`
	ProductId             int64              `json:"product_id" bson:"product_id"`
	OldCommissionRate     float64            `json:"old_commission_rate" bson:"old_commission_rate"`
	NewCommissionRate     float64            `json:"new_commission_rate" bson:"new_commission_rate"`
	PayloadCommissionRate float64            `json:"payload_commission_rate" bson:"payload_commission_rate"`
	Version               int8               `json:"version" bson:"version"`
	CreatedAt             time.Time          `json:"created_at" bson:"created_at"`
}

type DryRunLogAnalyzer struct {
	Job
}

func (drla *DryRunLogAnalyzer) Run() {
	// 1 gün önceki tarihi hesapla
	yesterday := time.Now().AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, yesterday.Location())

	log.GetLogger().Info(fmt.Sprintf("Analyzing dry run logs for date: %s", yesterday.Format("2006-01-02")))

	// MongoDB collection'ına bağlan
	collection := drla.Mongo.Database("commission_service").Collection("dry_run_logs")

	// Önceki günün loglarını filtrele
	filter := bson.M{
		"created_at": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
	}

	// CSV dosya adını hazırla
	fileName := fmt.Sprintf("assets/commission/dry_run_differences_%s.csv", yesterday.Format("2006-01-02"))
	csvCreated := false

	// Header'ı ilk yazımda ekle
	headerWritten := false

	// Batch işlemi için değişkenler
	batchSize := 1000
	skip := 0
	var totalCount int
	var totalDifferentCount int

	for {
		// Batch halinde veri çek
		findOptions := options.Find()
		findOptions.SetLimit(int64(batchSize))
		findOptions.SetSkip(int64(skip))
		findOptions.SetSort(bson.D{{"created_at", 1}})

		ctx := context.Background()
		cursor, err := collection.Find(ctx, filter, findOptions)
		if err != nil {
			log.GetLogger().Error("MongoDB query error", err)
			return
		}

		var batchLogs []DryRunLog
		var batchCount int

		// Bu batch'teki tüm logları oku
		for cursor.Next(ctx) {
			var dryRunLog DryRunLog
			if err := cursor.Decode(&dryRunLog); err != nil {
				log.GetLogger().Error("Document decode error", err)
				continue
			}
			batchLogs = append(batchLogs, dryRunLog)
			batchCount++
		}
		cursor.Close(ctx)

		// Eğer bu batch'te veri yoksa döngüden çık
		if batchCount == 0 {
			break
		}

		totalCount += batchCount

		// Bu batch'teki farklı logları topla
		var differentLogs [][]string

		// İlk batch'te header ekle
		if !headerWritten {
			header := []string{
				"ID",
				"Order ID",
				"Product ID",
				"Elasticsearch Commission Rate",
				"New Commission Rate",
				"Payload Commission Rate",
				"Version",
				"Created At",
				"Difference Type",
			}
			differentLogs = append(differentLogs, header)
			headerWritten = true
		}

		// Farklı olanları tespit et ve topla
		batchDifferentCount := 0
		for _, dryRunLog := range batchLogs {
			if drla.isCommissionRateDifferent(dryRunLog) {
				csvRow := drla.createCsvRow(dryRunLog)
				differentLogs = append(differentLogs, csvRow)
				batchDifferentCount++
			}
		}

		// Eğer farklı log varsa CSV'ye yaz
		if batchDifferentCount > 0 {
			helpers.Write(fileName, differentLogs)
			csvCreated = true
			totalDifferentCount += batchDifferentCount
		}

		log.GetLogger().Info(fmt.Sprintf("Batch processed: %d records, Different: %d, Total processed: %d", batchCount, batchDifferentCount, totalCount))

		// Eğer bu batch tam dolu değilse (son batch), döngüden çık
		if batchCount < batchSize {
			break
		}

		skip += batchSize
		batchLogs = nil // Memory temizle
	}

	// Eğer CSV oluşturduysan FTP'ye upload et ve mail gönder
	if csvCreated {
		log.GetLogger().Info(fmt.Sprintf("CSV created with %d different records, uploading to FTP", totalDifferentCount))
		drla.uploadToFtp(fileName, yesterday)
		drla.sendNotificationEmail(fileName, totalDifferentCount, yesterday)
	} else {
		log.GetLogger().Info("No different records found, no CSV created")
	}

	log.GetLogger().Info(fmt.Sprintf("Analysis completed - Total logs scanned: %d, Different logs found: %d", totalCount, totalDifferentCount))
}

// Mail gönderme fonksiyonu
func (drla *DryRunLogAnalyzer) sendNotificationEmail(filePath string, recordCount int, date time.Time) {
	// Config'den mail ayarlarını al
	cfg := config.GetConfig()

	// Insider mail client oluştur
	mailConfig := mail_client.InsiderMailConfig{
		Endpoint: cfg.InsiderMail.Endpoint,
		AuthKey:  cfg.InsiderMail.AuthKey,
		Timeout:  cfg.InsiderMail.Timeout,
	}

	mailClient := mail_client.NewInsiderMailClient(mailConfig)

	// Notification email listesini al
	emailList := cfg.NotificationEmails.DryRunReports
	if len(emailList) == 0 {
		log.GetLogger().Info("No notification emails configured, skipping email notification")
		return
	}

	log.GetLogger().Info(fmt.Sprintf("Sending notification email to %d recipients", len(emailList)))

	// Recipients listesini oluştur
	var recipients []mail_client.Recipient
	for _, email := range emailList {
		recipients = append(recipients, mail_client.Recipient{
			Email: email,
			Name:  "", // Name is optional
		})
	}

	// Template engine oluştur
	templateEngine := mail_client.NewTemplateEngine("")

	// Template data oluştur
	templateData := mail_client.CreateDryRunReportData(filePath, recordCount, date)

	// Template'den mail message oluştur - UniqueArgs'ı kaldır
	message, err := templateEngine.CreateMailFromTemplate(
		recipients,
		"destek@pttavm.com",
		"PTTAVM System",
		"dry_run_report",
		templateData,
		nil,
	)
	if err != nil {
		log.GetLogger().Error("Failed to create mail from template", err)
		return
	}

	// Mail gönder
	response, err := mailClient.SendMail(message)
	if err != nil {
		log.GetLogger().Error("Failed to send email", err)
		return
	}

	log.GetLogger().Info(fmt.Sprintf("Email sent successfully to %d recipients. Message ID: %s", len(recipients), response.MessageID))
}

// CSV satırı oluştur
func (drla *DryRunLogAnalyzer) createCsvRow(dryRunLog DryRunLog) []string {
	differenceType := drla.getDifferenceType(dryRunLog)

	return []string{
		dryRunLog.Id.Hex(),
		dryRunLog.OrderId,
		fmt.Sprintf("%d", dryRunLog.ProductId),
		fmt.Sprintf("%.2f", dryRunLog.OldCommissionRate),
		fmt.Sprintf("%.2f", dryRunLog.NewCommissionRate),
		fmt.Sprintf("%.2f", dryRunLog.PayloadCommissionRate),
		fmt.Sprintf("%d", dryRunLog.Version),
		dryRunLog.CreatedAt.Format("2006-01-02 15:04:05"),
		differenceType,
	}
}

// CSV dosyasını FTP'ye upload et
func (drla *DryRunLogAnalyzer) uploadToFtp(localFilePath string, date time.Time) {
	// Config'den FTP ayarlarını al
	cfg := config.GetConfig()
	ftpConfig := ftp_client.FtpConfig{
		Host:     cfg.FtpServer.Host,
		Port:     cfg.FtpServer.Port,
		Username: cfg.FtpServer.Username,
		Password: cfg.FtpServer.Password,
		Timeout:  cfg.FtpServer.Timeout,
	}

	// FTP client oluştur
	ftpClient := ftp_client.NewFtpClient(ftpConfig)

	// FTP'ye bağlan
	err := ftpClient.Connect()
	if err != nil {
		log.GetLogger().Error("FTP connection failed", err)
		return
	}
	defer ftpClient.Disconnect()

	// Remote dosya yolunu oluştur
	fileName := filepath.Base(localFilePath)
	remoteDir := fmt.Sprintf("/reports/dry_run_logs/%d/%02d", date.Year(), date.Month())
	remotePath := fmt.Sprintf("%s/%s", remoteDir, fileName)

	log.GetLogger().Info(fmt.Sprintf("Starting FTP upload: %s -> %s", localFilePath, remotePath))

	// Retry ile upload et
	err = ftpClient.UploadFileWithRetry(localFilePath, remotePath, 3)
	if err != nil {
		log.GetLogger().Error("FTP upload failed", err)
		return
	}

	log.GetLogger().Info(fmt.Sprintf("FTP upload completed successfully: %s", remotePath))

	// Upload sonrası dosya boyutunu kontrol et (opsiyonel)
	remoteSize, err := ftpClient.GetFileSize(remotePath)
	if err == nil {
		log.GetLogger().Info(fmt.Sprintf("Remote file size: %d bytes", remoteSize))
	}

	// FTP upload başarılı olduktan sonra local dosyayı sil
	err = os.Remove(localFilePath)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to delete local file after FTP upload: %s", localFilePath), err)
	} else {
		log.GetLogger().Info(fmt.Sprintf("Local file deleted successfully after FTP upload: %s", localFilePath))
	}
}

// Komisyon oranlarının farklı olup olmadığını kontrol et
func (drla *DryRunLogAnalyzer) isCommissionRateDifferent(dryRunLog DryRunLog) bool {
	// Old ve New commission rate'leri karşılaştır
	if dryRunLog.OldCommissionRate != dryRunLog.NewCommissionRate {
		return true
	}

	// New ve Payload commission rate'leri karşılaştır
	if dryRunLog.NewCommissionRate != dryRunLog.PayloadCommissionRate {
		return true
	}

	// Old ve Payload commission rate'leri karşılaştır
	if dryRunLog.OldCommissionRate != dryRunLog.PayloadCommissionRate {
		return true
	}

	return false
}

// Fark türünü belirle
func (drla *DryRunLogAnalyzer) getDifferenceType(dryRunLog DryRunLog) string {
	var differences []string

	if dryRunLog.OldCommissionRate != dryRunLog.NewCommissionRate {
		differences = append(differences, "Old-New")
	}

	if dryRunLog.NewCommissionRate != dryRunLog.PayloadCommissionRate {
		differences = append(differences, "New-Payload")
	}

	if dryRunLog.OldCommissionRate != dryRunLog.PayloadCommissionRate {
		differences = append(differences, "Old-Payload")
	}

	if len(differences) == 0 {
		return "None"
	}

	// Farklılıkları birleştir
	result := ""
	for i, diff := range differences {
		if i > 0 {
			result += ", "
		}
		result += diff
	}

	return result
}
