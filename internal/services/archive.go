package services

import (
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"

	"github.com/ab-dauletkhan/doozip/internal/entities"
	"github.com/ab-dauletkhan/doozip/internal/repositories"
)

var (
	ErrInvalidMimeType   = errors.New("invalid mime type")
	ErrEmptyFilesList    = errors.New("files list is empty")
	ErrNilFile           = errors.New("file is nil")
	ErrRepositoryNil     = errors.New("archive repository is nil")
	ErrInvalidArchiveZip = errors.New("invalid zip archive")
)

// ArchiveService defines the interface for archive operations at service level
type ArchiveService interface {
	GetArchiveInformation(file multipart.File, filename string) (*entities.ArchiveInfo, error)
	CreateZipArchive(files []*entities.FileData, archiveName string) (*entities.FileData, error)
	ValidateFiles(files []*entities.FileData) error
}

type archiveServiceImpl struct {
	archiveRepo repositories.ArchiveRepository
	log         *slog.Logger
}

// NewArchiveService creates a new instance of ArchiveService
func NewArchiveService(archiveRepo repositories.ArchiveRepository, log *slog.Logger) (ArchiveService, error) {
	if archiveRepo == nil {
		return nil, ErrRepositoryNil
	}

	if log == nil {
		log = slog.Default()
	}

	return &archiveServiceImpl{
		archiveRepo: archiveRepo,
		log:         log,
	}, nil
}

// GetArchiveInformation retrieves information about an archive file
func (s *archiveServiceImpl) GetArchiveInformation(file multipart.File, filename string) (*entities.ArchiveInfo, error) {
	const op = "archiveServiceImpl.GetArchiveInformation"

	if file == nil {
		return nil, fmt.Errorf("%s: %w", op, ErrNilFile)
	}

	if filename == "" {
		filename = "archive.zip"
	}

	archiveInfo, err := s.archiveRepo.GetArchiveInfo(file, filename)
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidZip) {
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidArchiveZip)
		}
		s.log.Error("failed to get archive info",
			"op", op,
			"error", err,
			"filename", filename,
		)
		return nil, fmt.Errorf("%s: failed to get archive info: %w", op, err)
	}

	return archiveInfo, nil
}

// CreateZipArchive creates a new zip archive from the provided files
func (s *archiveServiceImpl) CreateZipArchive(files []*entities.FileData, archiveName string) (*entities.FileData, error) {
	const op = "archiveServiceImpl.CreateZipArchive"

	if err := s.ValidateFiles(files); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if archiveName == "" {
		archiveName = "archive.zip"
	}

	buf, err := s.archiveRepo.CreateZipArchive(files)
	if err != nil {
		s.log.Error("failed to create zip archive",
			"op", op,
			"error", err,
			"filesCount", len(files),
		)
		return nil, fmt.Errorf("%s: failed to create zip archive: %w", op, err)
	}

	archiveFile := &entities.FileData{
		Name:     archiveName,
		Content:  buf.Bytes(),
		MIMEType: "application/zip",
	}

	if err := archiveFile.Validate(); err != nil {
		return nil, fmt.Errorf("%s: invalid archive file: %w", op, err)
	}

	return archiveFile, nil
}

// ValidateFiles validates a list of files for processing
func (s *archiveServiceImpl) ValidateFiles(files []*entities.FileData) error {
	const op = "archiveServiceImpl.ValidateFiles"

	if len(files) == 0 {
		return fmt.Errorf("%s: %w", op, ErrEmptyFilesList)
	}

	for _, file := range files {
		if file == nil {
			return fmt.Errorf("%s: file cannot be nil", op)
		}

		if err := file.Validate(); err != nil {
			return fmt.Errorf("%s: invalid file %s: %w", op, file.Name, err)
		}

		if !file.IsAllowedMimeType() {
			s.log.Warn("invalid mime type detected",
				"op", op,
				"filename", file.Name,
				"mimeType", file.MIMEType,
			)
			return fmt.Errorf("%s: %w: %s", op, ErrInvalidMimeType, file.MIMEType)
		}
	}

	return nil
}
