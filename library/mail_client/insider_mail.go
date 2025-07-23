package mail_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"my_toolbox/library/log"
	"net/http"
	"time"
)

type InsiderMailConfig struct {
	Endpoint string `json:"endpoint"`
	AuthKey  string `json:"auth_key"`
	Timeout  int    `json:"timeout"` // seconds
}

type InsiderMailClient struct {
	config InsiderMailConfig
	client *http.Client
}

// Insider API format structures
type Recipient struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type FromAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type ContentItem struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value"`
}

type ReplyTo struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type Attachment struct {
	Content  string `json:"content"`
	FileName string `json:"file_name"`
}

type Callback struct {
	Url    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

type MailMessage struct {
	Subject       string                 `json:"subject"`
	Tos           []Recipient            `json:"tos"`
	Name          string                 `json:"name,omitempty"`
	Email         string                 `json:"email,omitempty"`
	From          FromAddress            `json:"from"`
	Content       []ContentItem          `json:"content"`
	CC            []Recipient            `json:"cc,omitempty"`
	BCC           []Recipient            `json:"bcc,omitempty"`
	ReplyTo       *ReplyTo               `json:"reply_to,omitempty"`
	Attachments   []Attachment           `json:"attachments,omitempty"`
	DynamicFields map[string]interface{} `json:"dynamic_fields,omitempty"`
	UniqueArgs    map[string]interface{} `json:"unique_args,omitempty"`
	Callback      *Callback              `json:"callback,omitempty"`
}

type InsiderResponse struct {
	MessageID     string `json:"message_id"`
	StatusCode    int    `json:"status_code"`
	StatusMessage string `json:"status_message"`
}

func NewInsiderMailClient(config InsiderMailConfig) *InsiderMailClient {
	return &InsiderMailClient{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

func (imc *InsiderMailClient) SendMail(message MailMessage) (*InsiderResponse, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Debug: Print the full JSON being sent
	log.GetLogger().Info(fmt.Sprintf("Sending email request to Insider: %s", string(jsonData)))

	req, err := http.NewRequest("POST", imc.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-INS-AUTH-KEY", imc.config.AuthKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := imc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.GetLogger().Info(fmt.Sprintf("Insider API response: %s", string(body)))

	var insiderResp InsiderResponse
	if err := json.Unmarshal(body, &insiderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &insiderResp, fmt.Errorf("API error (status %d): %s", resp.StatusCode, insiderResp.StatusMessage)
	}

	return &insiderResp, nil
}

// Helper function to create a simple email
func CreateSimpleEmail(subject, toEmail, toName, fromEmail, fromName, htmlContent, textContent string) MailMessage {
	content := []ContentItem{
		{Type: "text/html", Value: htmlContent},
	}

	if textContent != "" {
		content = append(content, ContentItem{Type: "text/plain", Value: textContent})
	}

	return MailMessage{
		Subject: subject,
		Tos: []Recipient{
			{Name: toName, Email: toEmail},
		},
		From: FromAddress{
			Name:  fromName,
			Email: fromEmail,
		},
		Content: content,
	}
}

// Helper function to create email with multiple recipients
func CreateMultiRecipientEmail(subject string, recipients []Recipient, fromEmail, fromName, htmlContent, textContent string) MailMessage {
	content := []ContentItem{
		{Type: "text/html", Value: htmlContent},
	}

	if textContent != "" {
		content = append(content, ContentItem{Type: "text/plain", Value: textContent})
	}

	return MailMessage{
		Subject: subject,
		Tos:     recipients,
		From: FromAddress{
			Name:  fromName,
			Email: fromEmail,
		},
		Content: content,
	}
}
