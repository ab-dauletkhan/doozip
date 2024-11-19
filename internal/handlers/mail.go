package handlers

import (
	"fmt"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ab-dauletkhan/doozip/internal/services"
)

// MailHandler handles mail-related operations.
type MailHandler struct {
	service services.MailService
	log     *slog.Logger
}

// NewMailHandler creates a new MailHandler instance.
func NewMailHandler(svc services.MailService, log *slog.Logger) *MailHandler {
	return &MailHandler{service: svc, log: log}
}

// SendMail handles the mail sending request.
func (h *MailHandler) SendMail(w http.ResponseWriter, r *http.Request) {
	const op = "MailHandler.SendMail"

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.logError(op, "failed to parse multipart form", err)
		WriteError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logError(op, "file is required", err)
		WriteError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	if err := h.validateFileType(fileHeader.Filename); err != nil {
		h.logError(op, "invalid file type", err)
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	mailList := h.getMailList(r.FormValue("emails"))
	if len(mailList) == 0 {
		h.logError(op, "emails are required", nil)
		WriteError(w, http.StatusBadRequest, "emails are required")
		return
	}

	content, err := h.readFileContent(file, fileHeader.Size)
	if err != nil {
		h.logError(op, "failed to read file", err)
		WriteError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	if err := h.service.SendMail(mailList, fileHeader.Filename, mime.TypeByExtension(filepath.Ext(fileHeader.Filename)), content); err != nil {
		h.logError(op, "failed to send mail", err)
		WriteError(w, http.StatusInternalServerError, "failed to send mail")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Emails sent successfully."})
}

func (h *MailHandler) logError(op, message string, err error) {
	if err != nil {
		h.log.Error(fmt.Sprintf("%s - %s: %v", op, message, err))
	} else {
		h.log.Error(fmt.Sprintf("%s - %s", op, message))
	}
}

func (h *MailHandler) validateFileType(filename string) error {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" || (mimeType != "application/pdf" && mimeType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
		return fmt.Errorf("invalid file type")
	}
	return nil
}

func (h *MailHandler) getMailList(emails string) []string {
	if emails == "" {
		return nil
	}
	return strings.Split(emails, ",")
}

func (h *MailHandler) readFileContent(file multipart.File, size int64) ([]byte, error) {
	content := make([]byte, size)
	if _, err := file.Read(content); err != nil {
		return nil, err
	}
	return content, nil
}
