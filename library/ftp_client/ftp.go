package ftp_client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/secsy/goftp"
	"my_toolbox/library/log"
)

type FtpConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"` // seconds
}

type FtpClient struct {
	config FtpConfig
	client *goftp.Client
}

// NewFtpClient creates a new FTP client instance
func NewFtpClient(config FtpConfig) *FtpClient {
	if config.Timeout == 0 {
		config.Timeout = 30 // default 30 seconds
	}

	return &FtpClient{
		config: config,
	}
}

// Connect establishes connection to FTP server
func (fc *FtpClient) Connect() error {
	ftpConfig := goftp.Config{
		User:               fc.config.Username,
		Password:           fc.config.Password,
		ConnectionsPerHost: 10,
		Timeout:            time.Duration(fc.config.Timeout) * time.Second,
		Logger:             os.Stderr,
	}

	address := fmt.Sprintf("%s:%d", fc.config.Host, fc.config.Port)

	client, err := goftp.DialConfig(ftpConfig, address)
	if err != nil {
		log.GetLogger().Error("FTP connection failed", err)
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	fc.client = client
	log.GetLogger().Info(fmt.Sprintf("FTP connected to %s", address))
	return nil
}

// Disconnect closes the FTP connection
func (fc *FtpClient) Disconnect() error {
	if fc.client != nil {
		err := fc.client.Close()
		fc.client = nil
		if err != nil {
			log.GetLogger().Error("FTP disconnect error", err)
			return err
		}
		log.GetLogger().Info("FTP disconnected")
	}
	return nil
}

// UploadFile uploads a file to the FTP server
func (fc *FtpClient) UploadFile(localFilePath, remoteFilePath string) error {
	if fc.client == nil {
		return fmt.Errorf("FTP connection not established")
	}

	// Open local file
	file, err := os.Open(localFilePath)
	if err != nil {
		log.GetLogger().Error("Failed to open local file", err)
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file info for logging
	fileInfo, err := file.Stat()
	if err != nil {
		log.GetLogger().Error("Failed to get file info", err)
		return fmt.Errorf("failed to get file info: %w", err)
	}

	log.GetLogger().Info(fmt.Sprintf("Starting upload: %s (%d bytes) -> %s", localFilePath, fileInfo.Size(), remoteFilePath))

	// Create remote directory if needed
	remoteDir := filepath.Dir(remoteFilePath)
	if remoteDir != "." && remoteDir != "/" {
		err = fc.createRemoteDir(remoteDir)
		if err != nil {
			log.GetLogger().Error("Failed to create remote directory", err)
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
	}

	// Upload file
	err = fc.client.Store(remoteFilePath, file)
	if err != nil {
		log.GetLogger().Error("FTP upload failed", err)
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.GetLogger().Info(fmt.Sprintf("File uploaded successfully: %s", remoteFilePath))
	return nil
}

// UploadFileWithRetry uploads a file with retry mechanism
func (fc *FtpClient) UploadFileWithRetry(localFilePath, remoteFilePath string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.GetLogger().Info(fmt.Sprintf("Upload attempt %d/%d for %s", attempt, maxRetries, localFilePath))

		err := fc.UploadFile(localFilePath, remoteFilePath)
		if err == nil {
			return nil
		}

		lastErr = err
		log.GetLogger().Error(fmt.Sprintf("Upload attempt %d failed", attempt), err)

		if attempt < maxRetries {
			// Reconnect for next attempt
			fc.Disconnect()
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff

			if connectErr := fc.Connect(); connectErr != nil {
				log.GetLogger().Error("Reconnection failed", connectErr)
				continue
			}
		}
	}

	return fmt.Errorf("upload failed after %d attempts: %w", maxRetries, lastErr)
}

// ListFiles lists files in remote directory
func (fc *FtpClient) ListFiles(remotePath string) ([]os.FileInfo, error) {
	if fc.client == nil {
		return nil, fmt.Errorf("FTP connection not established")
	}

	entries, err := fc.client.ReadDir(remotePath)
	if err != nil {
		log.GetLogger().Error("Failed to list remote directory", err)
		return nil, fmt.Errorf("failed to list remote directory: %w", err)
	}

	return entries, nil
}

// DeleteFile deletes a file from FTP server
func (fc *FtpClient) DeleteFile(remoteFilePath string) error {
	if fc.client == nil {
		return fmt.Errorf("FTP connection not established")
	}

	err := fc.client.Delete(remoteFilePath)
	if err != nil {
		log.GetLogger().Error("Failed to delete remote file", err)
		return fmt.Errorf("failed to delete remote file: %w", err)
	}

	log.GetLogger().Info(fmt.Sprintf("File deleted: %s", remoteFilePath))
	return nil
}

// DownloadFile downloads a file from FTP server
func (fc *FtpClient) DownloadFile(remoteFilePath, localFilePath string) error {
	if fc.client == nil {
		return fmt.Errorf("FTP connection not established")
	}

	// Create local file
	localFile, err := os.Create(localFilePath)
	if err != nil {
		log.GetLogger().Error("Failed to create local file", err)
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Download file
	err = fc.client.Retrieve(remoteFilePath, localFile)
	if err != nil {
		log.GetLogger().Error("Failed to download file", err)
		return fmt.Errorf("failed to download file: %w", err)
	}

	log.GetLogger().Info(fmt.Sprintf("File downloaded: %s -> %s", remoteFilePath, localFilePath))
	return nil
}

// createRemoteDir creates remote directory recursively
func (fc *FtpClient) createRemoteDir(remotePath string) error {
	// Normalize path separators
	remotePath = strings.ReplaceAll(remotePath, "\\", "/")

	// Split path into parts
	parts := strings.Split(remotePath, "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		// Try to create directory
		_, err := fc.client.Mkdir(currentPath)
		if err != nil {
			// Directory might already exist, try to change to it
			entries, listErr := fc.client.ReadDir(filepath.Dir(currentPath))
			if listErr != nil {
				continue // Skip if we can't list
			}

			// Check if directory exists
			dirExists := false
			for _, entry := range entries {
				if entry.Name() == part && entry.IsDir() {
					dirExists = true
					break
				}
			}

			if !dirExists {
				return fmt.Errorf("failed to create directory %s: %w", currentPath, err)
			}
		}
	}

	return nil
}

// GetFileSize gets the size of a remote file
func (fc *FtpClient) GetFileSize(remoteFilePath string) (int64, error) {
	if fc.client == nil {
		return 0, fmt.Errorf("FTP connection not established")
	}

	// Get file info
	remoteDir := filepath.Dir(remoteFilePath)
	fileName := filepath.Base(remoteFilePath)

	entries, err := fc.client.ReadDir(remoteDir)
	if err != nil {
		log.GetLogger().Error("Failed to read remote directory", err)
		return 0, fmt.Errorf("failed to read remote directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == fileName {
			return entry.Size(), nil
		}
	}

	return 0, fmt.Errorf("file not found: %s", remoteFilePath)
}

// IsConnected checks if FTP connection is active
func (fc *FtpClient) IsConnected() bool {
	if fc.client == nil {
		return false
	}

	// Try to list current directory to check connection
	_, err := fc.client.ReadDir(".")
	return err == nil
}

// ChangeDir changes the current working directory
func (fc *FtpClient) ChangeDir(remotePath string) error {
	if fc.client == nil {
		return fmt.Errorf("FTP connection not established")
	}

	// goftp doesn't have explicit ChangeDir, but we can use ReadDir to test
	_, err := fc.client.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	return nil
}

// GetCurrentDir gets the current working directory
func (fc *FtpClient) GetCurrentDir() (string, error) {
	if fc.client == nil {
		return "", fmt.Errorf("FTP connection not established")
	}

	// goftp doesn't have built-in PWD, return "." as default
	return ".", nil
}
