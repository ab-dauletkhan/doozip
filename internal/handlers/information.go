package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"doozip/internal/entities"
	"doozip/internal/services"
)

const (
	maxFileSize     = 10 << 20 // 10 MB
	maxTotalSize    = 50 << 20 // 50 MB
	defaultFileName = "archive.zip"
)

var (
	ErrFileSizeTooLarge    = errors.New("file size exceeds maximum allowed size")
	ErrTotalSizeTooLarge   = errors.New("total size exceeds maximum allowed size")
	ErrNoFiles             = errors.New("no files provided")
	ErrServiceNil          = errors.New("archive service is nil")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrFileProcessingError = errors.New("error processing file")
)

// ArchiveHandler handles HTTP requests for archive operations
type ArchiveHandler struct {
	service services.ArchiveService
	log     *slog.Logger
}

// NewArchiveHandler creates a new instance of ArchiveHandler
func NewArchiveHandler(svc services.ArchiveService, log *slog.Logger) (*ArchiveHandler, error) {
	if svc == nil {
		return nil, ErrServiceNil
	}

	if log == nil {
		log = slog.Default()
	}

	return &ArchiveHandler{
		service: svc,
		log:     log,
	}, nil
}

// GetInformation handles requests to get archive information
func (h *ArchiveHandler) GetInformation(w http.ResponseWriter, r *http.Request) {
	const op = "ArchiveHandler.GetInformation"

	if err := h.validateRequest(r, "multipart/form-data"); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.log.Error("failed to get form file",
			"op", op,
			"error", err,
		)
		h.writeErrorResponse(w, http.StatusBadRequest, errors.New("file is required"))
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		h.writeErrorResponse(w, http.StatusBadRequest, ErrFileSizeTooLarge)
		return
	}

	result, err := h.service.GetArchiveInformation(file, header.Filename)
	if err != nil {
		h.log.Error("failed to get archive information",
			"op", op,
			"error", err,
			"filename", header.Filename,
		)
		h.writeErrorResponse(w, http.StatusInternalServerError, errors.New("failed to process archive"))
		return
	}

	h.writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    result,
	})
}

// CreateArchive handles requests to create a new archive
func (h *ArchiveHandler) CreateArchive(w http.ResponseWriter, r *http.Request) {
	const op = "ArchiveHandler.CreateArchive"

	if err := h.validateRequest(r, "multipart/form-data"); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseMultipartForm(maxTotalSize); err != nil {
		h.log.Error("failed to parse multipart form",
			"op", op,
			"error", err,
		)
		h.writeErrorResponse(w, http.StatusBadRequest, errors.New("failed to parse request"))
		return
	}

	files, err := h.processUploadedFiles(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	zipFile, err := h.service.CreateZipArchive(files, defaultFileName)
	if err != nil {
		h.log.Error("failed to create zip archive",
			"op", op,
			"error", err,
			"filesCount", len(files),
		)
		h.writeErrorResponse(w, http.StatusInternalServerError, errors.New("failed to create archive"))
		return
	}

	h.writeFileResponse(w, zipFile)
}

// processUploadedFiles processes uploaded files and returns FileData slice
func (h *ArchiveHandler) processUploadedFiles(r *http.Request) ([]*entities.FileData, error) {
	formFiles := r.MultipartForm.File["files[]"]
	if len(formFiles) == 0 {
		return nil, ErrNoFiles
	}

	var totalSize int64
	files := make([]*entities.FileData, 0, len(formFiles))

	for _, fileHeader := range formFiles {
		totalSize += fileHeader.Size
		if totalSize > maxTotalSize {
			return nil, ErrTotalSizeTooLarge
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", fileHeader.Filename, err)
		}

		fileData := &entities.FileData{
			Name:     fileHeader.Filename,
			Content:  content,
			MIMEType: mime.TypeByExtension(filepath.Ext(fileHeader.Filename)),
		}

		if err := fileData.Validate(); err != nil {
			return nil, fmt.Errorf("invalid file %s: %w", fileHeader.Filename, err)
		}

		files = append(files, fileData)
	}

	return files, nil
}

// validateRequest validates the HTTP request
func (h *ArchiveHandler) validateRequest(r *http.Request, expectedContentType string) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, expectedContentType) {
		return ErrInvalidContentType
	}

	return nil
}

// writeJSONResponse writes a JSON response
func (h *ArchiveHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error("failed to encode JSON response",
			"error", err,
		)
	}
}

// writeErrorResponse writes an error response
func (h *ArchiveHandler) writeErrorResponse(w http.ResponseWriter, status int, err error) {
	response := Response{
		Success: false,
		Error:   err.Error(),
	}
	h.writeJSONResponse(w, status, response)
}

// writeFileResponse writes a file response
func (h *ArchiveHandler) writeFileResponse(w http.ResponseWriter, file *entities.FileData) {
	w.Header().Set("Content-Type", file.MIMEType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(file.Content)))

	if _, err := w.Write(file.Content); err != nil {
		h.log.Error("failed to write file response",
			"error", err,
			"filename", file.Name,
		)
	}
}
