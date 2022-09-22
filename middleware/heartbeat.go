package middleware

// Ported from Chi's heartbeat middleware
// Source: https://github.com/go-chi/chi/blob/master/middleware/heartbeat.go

import (
	"net/http"
	"strings"
)

// Heartbeat endpoint middleware useful to setting up a path like
// `/ping` that load balancers or uptime testing external services
// can make a request before hitting any routes. It's also convenient
// to place this above ACL middlewares as well.
func Heartbeat(endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if (req.Method == http.MethodGet || req.Method == http.MethodHead) {
				if strings.EqualFold(req.URL.Path, endpoint) {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("."))
					return
				}
			}
			next.ServeHTTP(w, req)
		})
	}
}
