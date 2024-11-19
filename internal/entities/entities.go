package entities

import (
	"errors"
	"fmt"
	"mime"
	"path/filepath"
)

var (
	ErrEmptyFilename    = errors.New("filename cannot be empty")
	ErrInvalidFileSize  = errors.New("file size cannot be negative")
	ErrEmptyFiles       = errors.New("files list cannot be empty")
	ErrInvalidMimeType  = errors.New("invalid mime type")
	ErrContentRequired  = errors.New("file content is required")
	ErrFilepathRequired = errors.New("file path is required")
)

// AllowedMimeTypes contains the mime types that are allowed for file operations
var AllowedMimeTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/xml": true,
	"image/jpeg":      true,
	"image/png":       true,
	"application/pdf": true,
}

// ArchiveInfo represents detailed information about an archive and its contents
type ArchiveInfo struct {
	Filename    string        `json:"filename"`
	ArchiveSize int64         `json:"archive_size"`
	TotalSize   int64         `json:"total_size"`
	TotalFiles  uint          `json:"total_files"`
	Files       []FileDetails `json:"files"`
}

// Validate checks if the ArchiveInfo instance is valid
func (a *ArchiveInfo) Validate() error {
	if a.Filename == "" {
		return ErrEmptyFilename
	}
	if a.ArchiveSize < 0 {
		return fmt.Errorf("%w: archive size", ErrInvalidFileSize)
	}
	if a.TotalSize < 0 {
		return fmt.Errorf("%w: total size", ErrInvalidFileSize)
	}
	if a.TotalFiles < 0 {
		return fmt.Errorf("%w: total files", ErrInvalidFileSize)
	}
	if len(a.Files) == 0 {
		return ErrEmptyFiles
	}
	for _, file := range a.Files {
		if err := file.Validate(); err != nil {
			return fmt.Errorf("invalid file detail: %w", err)
		}
	}
	return nil
}

// CalculateTotals updates the total size and files count
func (a *ArchiveInfo) CalculateTotals() {
	var totalSize int64
	for _, file := range a.Files {
		totalSize += file.Size
	}
	a.TotalSize = totalSize
	a.TotalFiles = uint(len(a.Files))
}

// FileDetails contains information about a single file within an archive
type FileDetails struct {
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimetype"`
}

// Validate checks if the FileDetails instance is valid
func (f *FileDetails) Validate() error {
	if f.FilePath == "" {
		return ErrFilepathRequired
	}
	if f.Size < 0 {
		return fmt.Errorf("%w: file size", ErrInvalidFileSize)
	}
	if f.MimeType == "" {
		return ErrInvalidMimeType
	}
	return nil
}

// IsAllowedMimeType checks if the file's mime type is in the allowed list
func (f *FileDetails) IsAllowedMimeType() bool {
	return AllowedMimeTypes[f.MimeType]
}

// FileData represents a file's content and metadata
type FileData struct {
	Name     string
	Content  []byte
	MIMEType string
}

// Validate checks if the FileData instance is valid
func (f *FileData) Validate() error {
	if f.Name == "" {
		return ErrEmptyFilename
	}
	if len(f.Content) == 0 {
		return ErrContentRequired
	}
	if f.MIMEType == "" {
		// Try to detect MIME type from file extension
		ext := filepath.Ext(f.Name)
		if mtype := mime.TypeByExtension(ext); mtype != "" {
			f.MIMEType = mtype
		} else {
			return ErrInvalidMimeType
		}
	}
	return nil
}

// IsAllowedMimeType checks if the file's mime type is in the allowed list
func (f *FileData) IsAllowedMimeType() bool {
	return AllowedMimeTypes[f.MIMEType]
}

// Size returns the size of the file content in bytes
func (f *FileData) Size() int64 {
	return int64(len(f.Content))
}
