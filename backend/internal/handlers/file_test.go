package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

func newFileHandler(t *testing.T) (*FileHandler, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("UPLOAD_DIR", dir)
	return NewFileHandler(), dir
}

func multipartRequest(t *testing.T, field, filename, contentType string, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, field, filename))
	if contentType != "" {
		hdr.Set("Content-Type", contentType)
	}
	part, err := mw.CreatePart(hdr)
	if err != nil {
		t.Fatalf("Failed to create multipart part: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("Failed to write multipart data: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestFileHandler_Upload(t *testing.T) {
	handler, dir := newFileHandler(t)

	req := multipartRequest(t, "file", "photo.png", "image/png", []byte("fake png bytes"))
	w := httptest.NewRecorder()

	handler.HandleUpload(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d (body %q)", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	filename, _ := resp["filename"].(string)
	if filename == "" {
		t.Fatalf("Expected a filename in response, got %v", resp)
	}
	if url, _ := resp["url"].(string); url != "/uploads/"+filename {
		t.Errorf("Expected url '/uploads/%s', got %q", filename, url)
	}
	if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
		t.Errorf("Expected uploaded file on disk: %v", err)
	}
}

func TestFileHandler_Upload_RejectsType(t *testing.T) {
	handler, _ := newFileHandler(t)

	req := multipartRequest(t, "file", "notes.txt", "text/plain", []byte("hello"))
	w := httptest.NewRecorder()

	handler.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for disallowed type, got %d", w.Code)
	}
}

func TestFileHandler_Upload_NoFile(t *testing.T) {
	handler, _ := newFileHandler(t)

	req := multipartRequest(t, "other", "photo.png", "image/png", []byte("x"))
	w := httptest.NewRecorder()

	handler.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 when no file field, got %d", w.Code)
	}
}

func TestFileHandler_Delete(t *testing.T) {
	handler, dir := newFileHandler(t)

	name := "todelete.png"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0o644); err != nil {
		t.Fatalf("Failed to seed file: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/upload/"+name, nil)
	req = withParams(req, map[string]string{"filename": name}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
		t.Errorf("Expected file to be removed, stat err: %v", err)
	}
}

func TestFileHandler_Delete_InvalidName(t *testing.T) {
	handler, _ := newFileHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/upload/x", nil)
	req = withParams(req, map[string]string{"filename": "../secret.txt"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for path traversal, got %d", w.Code)
	}
}

func TestFileHandler_Delete_NotFound(t *testing.T) {
	handler, _ := newFileHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/upload/ghost.png", nil)
	req = withParams(req, map[string]string{"filename": "ghost.png"}, testOwner)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestFileHandler_ServeFile(t *testing.T) {
	handler, dir := newFileHandler(t)

	name := "served.png"
	content := []byte("image-bytes")
	if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
		t.Fatalf("Failed to seed file: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/uploads/"+name, nil)
	req = withParams(req, map[string]string{"filename": name}, "")
	w := httptest.NewRecorder()

	handler.ServeFile(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Expected content-type image/png, got %q", ct)
	}
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Errorf("Served body did not match file content")
	}
}

func TestFileHandler_ServeFile_InvalidName(t *testing.T) {
	handler, _ := newFileHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/uploads/x", nil)
	req = withParams(req, map[string]string{"filename": "../etc/passwd"}, "")
	w := httptest.NewRecorder()

	handler.ServeFile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for path traversal, got %d", w.Code)
	}
}
