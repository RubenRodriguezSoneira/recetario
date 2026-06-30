package appmiddleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHandler_ClientErrorsExposeUIDescription(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		description string
	}{
		{"bad request", http.StatusBadRequest, "La solicitud no es válida. Revisa los datos e inténtalo nuevamente."},
		{"unauthorized", http.StatusUnauthorized, "Debes iniciar sesión para continuar."},
		{"forbidden", http.StatusForbidden, "No tienes permisos para realizar esta acción."},
		{"conflict", http.StatusConflict, "No se pudo completar la operación por un conflicto de datos."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "client error", tt.status)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ErrorHandler(next).ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}

			var payload ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if payload.Description != tt.description {
				t.Fatalf("description = %q, want %q", payload.Description, tt.description)
			}
			if payload.UserMessage != payload.Description {
				t.Fatalf("user_message = %q, want %q", payload.UserMessage, payload.Description)
			}
		})
	}
}
