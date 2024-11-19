package repositories

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"path/filepath"

	"github.com/ab-dauletkhan/doozip/internal/entities"
)

var (
	ErrEmptyFile      = errors.New("file is empty")
	ErrInvalidZip     = errors.New("invalid zip file")
	ErrEmptyFilesList = errors.New("files list is empty")
)

// ArchiveRepository defines the interface for archive operations
type ArchiveRepository interface {
	GetArchiveInfo(file multipart.File, filename string) (*entities.ArchiveInfo, error)
	CreateZipArchive(files []*entities.FileData) (*bytes.Buffer, error)
}

type archiveRepositoryImpl struct {
	log *slog.Logger
}

// NewArchiveRepository creates a new instance of ArchiveRepository
func NewArchiveRepository(log *slog.Logger) ArchiveRepository {
	return &archiveRepositoryImpl{log: log}
}

// GetArchiveInfo extracts and returns information about a zip archive
func (r *archiveRepositoryImpl) GetArchiveInfo(file multipart.File, filename string) (*entities.ArchiveInfo, error) {
	const op = "archiveRepositoryImpl.GetArchiveInfo"

	if file == nil {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyFile)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		r.log.Error("failed to read file content",
			"op", op,
			"error", err,
		)
		return nil, fmt.Errorf("%s: failed to read file: %w", op, err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyFile)
	}

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		r.log.Error("failed to create zip reader",
			"op", op,
			"error", err,
		)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidZip)
	}

	archiveInfo := &entities.ArchiveInfo{
		Filename:    filename,
		ArchiveSize: int64(len(content)),
		Files:       make([]entities.FileDetails, 0, len(reader.File)),
	}

	if err := r.processZipFiles(reader, archiveInfo); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	archiveInfo.CalculateTotals()

	if err := archiveInfo.Validate(); err != nil {
		return nil, fmt.Errorf("%s: invalid archive info: %w", op, err)
	}

	return archiveInfo, nil
}

// processZipFiles processes files within the zip archive and populates archive info
func (r *archiveRepositoryImpl) processZipFiles(reader *zip.Reader, archiveInfo *entities.ArchiveInfo) error {
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		fileDetails := entities.FileDetails{
			FilePath: filepath.Clean(f.Name),
			Size:     f.FileInfo().Size(),
			MimeType: r.detectMimeType(f.Name),
		}

		if err := fileDetails.Validate(); err != nil {
			r.log.Warn("invalid file in archive",
				"filepath", fileDetails.FilePath,
				"error", err,
			)
			continue
		}

		archiveInfo.Files = append(archiveInfo.Files, fileDetails)
	}

	return nil
}

// CreateZipArchive creates a new zip archive from the provided files
func (r *archiveRepositoryImpl) CreateZipArchive(files []*entities.FileData) (*bytes.Buffer, error) {
	const op = "archiveRepositoryImpl.CreateZipArchive"

	if len(files) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyFilesList)
	}

	// Validate all files before processing
	for _, file := range files {
		if err := file.Validate(); err != nil {
			return nil, fmt.Errorf("%s: invalid file %s: %w", op, file.Name, err)
		}
	}

	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	defer func() {
		if err := writer.Close(); err != nil {
			r.log.Error("failed to close zip writer",
				"op", op,
				"error", err,
			)
		}
	}()

	for _, file := range files {
		if err := r.addFileToZip(writer, file); err != nil {
			return nil, fmt.Errorf("%s: failed to add file %s: %w", op, file.Name, err)
		}
	}

	return buf, nil
}

// addFileToZip adds a single file to the zip archive
func (r *archiveRepositoryImpl) addFileToZip(writer *zip.Writer, file *entities.FileData) error {
	w, err := writer.Create(filepath.Clean(file.Name))
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %w", err)
	}

	if _, err := w.Write(file.Content); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// detectMimeType attempts to detect the MIME type of a file
func (r *archiveRepositoryImpl) detectMimeType(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}
