package repositories

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/ab-dauletkhan/doozip/internal/config"
	"github.com/ab-dauletkhan/doozip/internal/entities"
)

var (
	ErrInvalidSMTPConfig = errors.New("invalid SMTP configuration")
	ErrInvalidRecipients = errors.New("invalid recipients")
	ErrInvalidSubject    = errors.New("subject cannot be empty")
	ErrInvalidFile       = errors.New("invalid file data")
	ErrSMTPSendFailed    = errors.New("failed to send email")

	// Email validation regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// MailRepository defines the interface for email operations
type MailRepository interface {
	SendMail(to []string, subject, body string, file *entities.FileData) error
	ValidateConfig() error
}

// MailRepositoryImpl implements the MailRepository interface
type MailRepositoryImpl struct {
	smtpHost string
	smtpPort string
	username string
	password string
	auth     smtp.Auth
}

// NewMailRepository creates a new instance of MailRepositoryImpl with validation
func NewMailRepository(cfg *config.SMTP) (*MailRepositoryImpl, error) {
	if cfg == nil {
		return nil, fmt.Errorf("%w: configuration is nil", ErrInvalidSMTPConfig)
	}

	repo := &MailRepositoryImpl{
		smtpHost: cfg.Host,
		smtpPort: cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
	}

	if err := repo.ValidateConfig(); err != nil {
		return nil, err
	}

	// Initialize SMTP auth
	repo.auth = smtp.PlainAuth("", repo.username, repo.password, repo.smtpHost)

	return repo, nil
}

// ValidateConfig checks if the SMTP configuration is valid
func (m *MailRepositoryImpl) ValidateConfig() error {
	if m.smtpHost == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidSMTPConfig)
	}
	if m.smtpPort == "" {
		return fmt.Errorf("%w: port is required", ErrInvalidSMTPConfig)
	}
	if m.username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidSMTPConfig)
	}
	if m.password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidSMTPConfig)
	}
	return nil
}

// validateEmails checks if all email addresses are valid
func validateEmails(emails []string) error {
	if len(emails) == 0 {
		return fmt.Errorf("%w: no recipients provided", ErrInvalidRecipients)
	}

	for _, email := range emails {
		if !emailRegex.MatchString(email) {
			return fmt.Errorf("%w: invalid email format: %s", ErrInvalidRecipients, email)
		}
	}
	return nil
}

// createEmailContent builds the email content with attachment
func (m *MailRepositoryImpl) createEmailContent(to []string, subject, body string, file *entities.FileData) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// Write email headers
	headers := map[string]string{
		"Subject":      subject,
		"To":           strings.Join(to, ","),
		"MIME-Version": "1.0",
	}

	for key, value := range headers {
		if _, err := fmt.Fprintf(buf, "%s: %s\r\n", key, value); err != nil {
			return nil, fmt.Errorf("failed to write header %s: %w", key, err)
		}
	}

	// Create multipart writer
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()

	if _, err := fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary); err != nil {
		return nil, fmt.Errorf("failed to write content type: %w", err)
	}

	// Write body
	if err := m.writeMessageBody(buf, boundary, body); err != nil {
		return nil, err
	}

	// Write attachment
	if err := m.writeAttachment(buf, boundary, file); err != nil {
		return nil, err
	}

	// Close boundary
	if _, err := fmt.Fprintf(buf, "--%s--", boundary); err != nil {
		return nil, fmt.Errorf("failed to close boundary: %w", err)
	}

	return buf, nil
}

// writeMessageBody writes the email body part
func (m *MailRepositoryImpl) writeMessageBody(buf *bytes.Buffer, boundary, body string) error {
	if _, err := fmt.Fprintf(buf, "--%s\r\n", boundary); err != nil {
		return fmt.Errorf("failed to write body boundary: %w", err)
	}
	if _, err := fmt.Fprintf(buf, "Content-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", body); err != nil {
		return fmt.Errorf("failed to write body content: %w", err)
	}
	return nil
}

// writeAttachment writes the file attachment part
func (m *MailRepositoryImpl) writeAttachment(buf *bytes.Buffer, boundary string, file *entities.FileData) error {
	if _, err := fmt.Fprintf(buf, "--%s\r\n", boundary); err != nil {
		return fmt.Errorf("failed to write attachment boundary: %w", err)
	}

	headers := map[string]string{
		"Content-Type":              file.MIMEType,
		"Content-Transfer-Encoding": "base64",
		"Content-Disposition":       fmt.Sprintf("attachment; filename=%s", file.Name),
	}

	for key, value := range headers {
		if _, err := fmt.Fprintf(buf, "%s: %s\r\n", key, value); err != nil {
			return fmt.Errorf("failed to write attachment header %s: %w", key, err)
		}
	}

	if _, err := buf.WriteString("\r\n"); err != nil {
		return fmt.Errorf("failed to write attachment separator: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(file.Content)
	if _, err := buf.WriteString(encoded); err != nil {
		return fmt.Errorf("failed to write attachment content: %w", err)
	}

	if _, err := buf.WriteString("\r\n"); err != nil {
		return fmt.Errorf("failed to write attachment ending: %w", err)
	}

	return nil
}

// SendMail sends an email with an attachment
func (m *MailRepositoryImpl) SendMail(to []string, subject, body string, file *entities.FileData) error {
	// Validate inputs
	if err := validateEmails(to); err != nil {
		return err
	}
	if subject == "" {
		return ErrInvalidSubject
	}
	if file == nil {
		return fmt.Errorf("%w: file is nil", ErrInvalidFile)
	}
	if err := file.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}

	// Create email content
	content, err := m.createEmailContent(to, subject, body, file)
	if err != nil {
		return fmt.Errorf("failed to create email content: %w", err)
	}

	// Send email
	err = smtp.SendMail(
		fmt.Sprintf("%s:%s", m.smtpHost, m.smtpPort),
		m.auth,
		m.username,
		to,
		content.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSMTPSendFailed, err)
	}

	return nil
}
