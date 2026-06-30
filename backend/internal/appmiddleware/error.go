package appmiddleware

import (
	"encoding/json"
	"net/http"
	"time"

	"recipe-app/internal/logger"

	"github.com/go-chi/chi/v5/middleware"
)

type ErrorResponse struct {
	Error       string `json:"error"`
	Message     string `json:"message,omitempty"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	UserMessage string `json:"user_message,omitempty"`
}

type AppError struct {
	StatusCode int
	Message    string
	Code       string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func NewAppError(statusCode int, message, code string, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Err:        err,
	}
}

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := r.Context()
				logger.LogError(ctx, nil, "Panic recovered")

				WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", "Ocurrió un error inesperado. Inténtalo nuevamente.")
			}
		}()

		next.ServeHTTP(&errorResponseWriter{ResponseWriter: w, r: r}, r)
	})
}

type errorResponseWriter struct {
	http.ResponseWriter
	r           *http.Request
	errorStatus int
}

func (erw *errorResponseWriter) WriteHeader(statusCode int) {
	if erw.errorStatus != 0 {
		return
	}

	if statusCode >= 400 {
		erw.errorStatus = statusCode
		ctx := erw.r.Context()
		logger.LogError(ctx, nil, "HTTP error response")

		writeJSONErrorResponse(erw.ResponseWriter, statusCode, ErrorResponse{
			Error:       http.StatusText(statusCode),
			Message:     getErrorMessage(statusCode),
			Code:        getErrorCode(statusCode),
			Description: getErrorDescription(statusCode),
			UserMessage: getErrorDescription(statusCode),
		})
		return
	}

	erw.ResponseWriter.WriteHeader(statusCode)
}

func (erw *errorResponseWriter) Write(p []byte) (int, error) {
	if erw.errorStatus >= 400 {
		return len(p), nil
	}
	return erw.ResponseWriter.Write(p)
}

func WriteJSONError(w http.ResponseWriter, statusCode int, code, message, description string) {
	if code == "" {
		code = getErrorCode(statusCode)
	}
	if message == "" {
		message = getErrorMessage(statusCode)
	}
	if description == "" {
		description = getErrorDescription(statusCode)
	}

	writeJSONErrorResponse(w, statusCode, ErrorResponse{
		Error:       http.StatusText(statusCode),
		Message:     message,
		Code:        code,
		Description: description,
		UserMessage: description,
	})
}

func writeJSONErrorResponse(w http.ResponseWriter, statusCode int, payload ErrorResponse) {
	target := w
	if erw, ok := w.(*errorResponseWriter); ok {
		target = erw.ResponseWriter
	}

	target.Header().Set("Content-Type", "application/json")
	target.WriteHeader(statusCode)
	_ = json.NewEncoder(target).Encode(payload)
}

func getErrorMessage(statusCode int) string {
	messages := map[int]string{
		http.StatusBadRequest:          "Invalid request parameters",
		http.StatusUnauthorized:        "Authentication required",
		http.StatusForbidden:           "Access forbidden",
		http.StatusNotFound:            "Resource not found",
		http.StatusMethodNotAllowed:    "Method not allowed",
		http.StatusTooManyRequests:     "Rate limit exceeded",
		http.StatusInternalServerError: "Internal server error",
		http.StatusBadGateway:          "Service unavailable",
		http.StatusServiceUnavailable:  "Service temporarily unavailable",
	}

	if msg, ok := messages[statusCode]; ok {
		return msg
	}
	return "An error occurred"
}

func getErrorCode(statusCode int) string {
	codes := map[int]string{
		http.StatusBadRequest:          "BAD_REQUEST",
		http.StatusUnauthorized:        "UNAUTHORIZED",
		http.StatusForbidden:           "FORBIDDEN",
		http.StatusNotFound:            "NOT_FOUND",
		http.StatusMethodNotAllowed:    "METHOD_NOT_ALLOWED",
		http.StatusTooManyRequests:     "RATE_LIMIT_EXCEEDED",
		http.StatusInternalServerError: "INTERNAL_ERROR",
		http.StatusBadGateway:          "BAD_GATEWAY",
		http.StatusServiceUnavailable:  "SERVICE_UNAVAILABLE",
	}

	if code, ok := codes[statusCode]; ok {
		return code
	}
	return "UNKNOWN_ERROR"
}

func getErrorDescription(statusCode int) string {
	descriptions := map[int]string{
		http.StatusBadRequest:          "La solicitud no es válida. Revisa los datos e inténtalo nuevamente.",
		http.StatusUnauthorized:        "Debes iniciar sesión para continuar.",
		http.StatusForbidden:           "No tienes permisos para realizar esta acción.",
		http.StatusNotFound:            "No encontramos el recurso solicitado.",
		http.StatusConflict:            "No se pudo completar la operación por un conflicto de datos.",
		http.StatusMethodNotAllowed:    "Método no permitido.",
		http.StatusTooManyRequests:     "Has realizado demasiadas solicitudes. Inténtalo en unos minutos.",
		http.StatusInternalServerError: "Ocurrió un error interno. Inténtalo más tarde.",
		http.StatusBadGateway:          "El servicio no está disponible en este momento.",
		http.StatusServiceUnavailable:  "El servicio está temporalmente no disponible.",
	}

	if description, ok := descriptions[statusCode]; ok {
		return description
	}
	return "No se pudo completar la solicitud."
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := logger.WithCorrelationID(r.Context())
		r = r.WithContext(ctx)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		logger.LogRequest(ctx, r.Method, r.RequestURI, duration, ww.Status())
	})
}
