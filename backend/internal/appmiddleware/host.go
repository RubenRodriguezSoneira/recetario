package appmiddleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// CanonicalLoopbackHost redirects loopback aliases (127.0.0.1 / ::1) to
// localhost so auth cookies remain consistent in local development.
func CanonicalLoopbackHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, port := splitRequestHost(r.Host)
		if !shouldRedirectToLocalhost(host) {
			next.ServeHTTP(w, r)
			return
		}

		targetHost := "localhost"
		if port != "" {
			targetHost = net.JoinHostPort(targetHost, port)
		}

		targetURL := &url.URL{
			Scheme:   requestScheme(r),
			Host:     targetHost,
			Path:     r.URL.Path,
			RawPath:  r.URL.RawPath,
			RawQuery: r.URL.RawQuery,
		}
		http.Redirect(w, r, targetURL.String(), http.StatusTemporaryRedirect)
	})
}

func splitRequestHost(hostPort string) (string, string) {
	host := hostPort
	port := ""

	if parsedHost, parsedPort, err := net.SplitHostPort(hostPort); err == nil {
		host = parsedHost
		port = parsedPort
	}

	return strings.Trim(host, "[]"), port
}

func shouldRedirectToLocalhost(host string) bool {
	switch strings.ToLower(host) {
	case "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func requestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); strings.EqualFold(forwardedProto, "https") {
		return "https"
	}
	return "http"
}
