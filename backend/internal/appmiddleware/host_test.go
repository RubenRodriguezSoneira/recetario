package appmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCanonicalLoopbackHost(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		path         string
		forwarded    string
		wantStatus   int
		wantLocation string
		wantNext     bool
	}{
		{
			name:         "redirects ipv4 loopback to localhost preserving port and query",
			host:         "127.0.0.1:8080",
			path:         "/recipes/1?tab=details",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "http://localhost:8080/recipes/1?tab=details",
		},
		{
			name:         "redirects ipv6 loopback to localhost",
			host:         "[::1]:8080",
			path:         "/api/recipes/1",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "http://localhost:8080/api/recipes/1",
		},
		{
			name:       "does not redirect localhost",
			host:       "localhost:8080",
			path:       "/recipes/1",
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:       "does not redirect non-loopback host",
			host:       "example.com",
			path:       "/recipes/1",
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:         "respects forwarded https scheme in redirect",
			host:         "127.0.0.1:8080",
			path:         "/",
			forwarded:    "https",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "https://localhost:8080/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "http://"+tt.host+tt.path, nil)
			req.Host = tt.host
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-Proto", tt.forwarded)
			}
			w := httptest.NewRecorder()

			CanonicalLoopbackHost(next).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantLocation != "" {
				if got := w.Header().Get("Location"); got != tt.wantLocation {
					t.Fatalf("expected location %q, got %q", tt.wantLocation, got)
				}
			}
			if nextCalled != tt.wantNext {
				t.Fatalf("expected next called=%v, got %v", tt.wantNext, nextCalled)
			}
		})
	}
}
