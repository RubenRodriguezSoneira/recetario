package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/logger"
)

// FileHandler serves image uploads from the local filesystem. It has no
// database dependency: files are stored under uploadDir and exposed via
// ServeFile. The handler validates content types and guards against path
// traversal so user-supplied filenames cannot escape the upload directory.
type FileHandler struct {
	uploadDir    string
	maxFileSize  int64
	allowedTypes []string
}

func NewFileHandler() *FileHandler {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	return &FileHandler{
		uploadDir:    uploadDir,
		maxFileSize:  10 * 1024 * 1024, // 10MB
		allowedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
	}
}

// HandleUpload stores a single uploaded image and returns its public URL.
func (h *FileHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Processing file upload")

	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
		logger.LogError(r.Context(), err, "Failed to parse multipart form")
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !h.isAllowedType(contentType) {
		log.Info("Rejected file upload", "type", contentType)
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	filename := h.generateFilename(header.Filename)
	if err := h.saveFile(file, filename); err != nil {
		logger.LogError(r.Context(), err, "Failed to save uploaded file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("File uploaded successfully", "filename", filename, "size", header.Size)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":      "File uploaded successfully",
		"filename":     filename,
		"url":          "/uploads/" + filename,
		"size":         header.Size,
		"content_type": contentType,
	})
}

// HandleMultiUpload stores up to five uploaded images, skipping any that are
// not allowed image types.
func (h *FileHandler) HandleMultiUpload(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Processing multiple file upload")

	if err := r.ParseMultipartForm(h.maxFileSize * 5); err != nil {
		logger.LogError(r.Context(), err, "Failed to parse multipart form")
		http.Error(w, "Files too large", http.StatusRequestEntityTooLarge)
		return
	}

	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	const maxFiles = 5
	uploaded := make([]map[string]interface{}, 0, maxFiles)

	for fieldName, fileHeaders := range r.MultipartForm.File {
		for _, header := range fileHeaders {
			if len(uploaded) >= maxFiles {
				break
			}

			contentType := header.Header.Get("Content-Type")
			if !h.isAllowedType(contentType) {
				continue
			}

			file, err := header.Open()
			if err != nil {
				logger.LogError(r.Context(), err, "Failed to open uploaded file")
				continue
			}

			filename := h.generateFilename(header.Filename)
			err = h.saveFile(file, filename)
			file.Close()
			if err != nil {
				logger.LogError(r.Context(), err, "Failed to save uploaded file")
				continue
			}

			uploaded = append(uploaded, map[string]interface{}{
				"field_name":    fieldName,
				"filename":      filename,
				"original_name": header.Filename,
				"url":           "/uploads/" + filename,
				"size":          header.Size,
				"content_type":  contentType,
			})
		}
	}

	log.Info("Files uploaded successfully", "count", len(uploaded))
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Files uploaded successfully",
		"files":   uploaded,
		"count":   len(uploaded),
	})
}

// HandleDelete removes a previously uploaded file.
func (h *FileHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if !validUploadName(filename) {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(filePath); err != nil {
		logger.LogError(r.Context(), err, "Failed to delete file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.FromContext(r.Context()).Info("File deleted successfully", "filename", filename)
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "File deleted successfully"})
}

// ServeFile streams an uploaded file with an image content type.
func (h *FileHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if !validUploadName(filename) {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if ct := contentTypeForExt(filePath); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	http.ServeFile(w, r, filePath)
}

func (h *FileHandler) isAllowedType(contentType string) bool {
	for _, allowed := range h.allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// generateFilename derives a collision-resistant name from the original
// filename and the current time. The hash is used only for naming, not
// security.
func (h *FileHandler) generateFilename(original string) string {
	ext := filepath.Ext(original)
	hash := md5.Sum([]byte(fmt.Sprintf("%s%d", original, time.Now().UnixNano())))
	return fmt.Sprintf("%x%s", hash, ext)
}

func (h *FileHandler) saveFile(src io.Reader, filename string) error {
	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	dst, err := os.Create(filepath.Join(h.uploadDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// validUploadName rejects empty names and any path-traversal attempt so a
// request cannot read or delete files outside the upload directory.
func validUploadName(name string) bool {
	if name == "" {
		return false
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		return false
	}
	return true
}

func contentTypeForExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}
