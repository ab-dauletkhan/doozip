package services

import (
	"errors"
	"fmt"

	"github.com/ab-dauletkhan/doozip/internal/entities"
	"github.com/ab-dauletkhan/doozip/internal/repositories"
)

var (
	ErrNoRecipients   = errors.New("no recipients provided")
	ErrInvalidFile    = errors.New("invalid file data")
	ErrMailSendFailed = errors.New("failed to send mail")
)

// MailService defines the interface for mail operations
type MailService interface {
	// SendMail sends a file to multiple recipients
	SendMail(to []string, filename, mimeType string, fileContent []byte) error
	// SendMailWithTemplate sends a file with custom subject and body template
	SendMailWithTemplate(to []string, filename, mimeType string, fileContent []byte, subject, bodyTemplate string) error
	// ValidateFileType checks if the given mime type is supported
	ValidateFileType(mimeType string) error
}

// MailServiceImpl implements the MailService interface
type MailServiceImpl struct {
	repo repositories.MailRepository
}

// NewMailService creates a new instance of MailService with validation
func NewMailService(repo repositories.MailRepository) (MailService, error) {
	if repo == nil {
		return nil, errors.New("mail repository is required")
	}

	return &MailServiceImpl{
		repo: repo,
	}, nil
}

// validateInput checks if the input parameters are valid
func (s *MailServiceImpl) validateInput(to []string, filename, mimeType string, fileContent []byte) error {
	if len(to) == 0 {
		return ErrNoRecipients
	}

	if filename == "" {
		return fmt.Errorf("%w: filename is required", ErrInvalidFile)
	}

	if len(fileContent) == 0 {
		return fmt.Errorf("%w: file content is empty", ErrInvalidFile)
	}

	if err := s.ValidateFileType(mimeType); err != nil {
		return err
	}

	return nil
}

// ValidateFileType checks if the given mime type is supported
func (s *MailServiceImpl) ValidateFileType(mimeType string) error {
	allowedTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/pdf": true,
	}

	if !allowedTypes[mimeType] {
		return fmt.Errorf("%w: %s", ErrInvalidMimeType, mimeType)
	}

	return nil
}

// createFileData creates a new FileData instance with validation
func (s *MailServiceImpl) createFileData(filename, mimeType string, fileContent []byte) (*entities.FileData, error) {
	fileData := &entities.FileData{
		Name:     filename,
		Content:  fileContent,
		MIMEType: mimeType,
	}

	if err := fileData.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}

	return fileData, nil
}

// SendMail sends a file to multiple recipients with default subject and body
func (s *MailServiceImpl) SendMail(to []string, filename, mimeType string, fileContent []byte) error {
	return s.SendMailWithTemplate(
		to,
		filename,
		mimeType,
		fileContent,
		"File Attachment",
		"Please find the attached file.",
	)
}

// SendMailWithTemplate sends a file with custom subject and body template
func (s *MailServiceImpl) SendMailWithTemplate(to []string, filename, mimeType string, fileContent []byte, subject, bodyTemplate string) error {
	// Validate input parameters
	if err := s.validateInput(to, filename, mimeType, fileContent); err != nil {
		return err
	}

	// Create and validate file data
	fileData, err := s.createFileData(filename, mimeType, fileContent)
	if err != nil {
		return err
	}

	// Use the repository to send the email
	if err := s.repo.SendMail(to, subject, bodyTemplate, fileData); err != nil {
		return fmt.Errorf("%w: %v", ErrMailSendFailed, err)
	}

	return nil
}
