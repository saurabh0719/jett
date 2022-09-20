package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
)

// BasicAuth middleware - Implements middleware handler
// RFC 2617, Section 2. (https://www.rfc-editor.org/rfc/rfc2617.html#section-2)
// Ref - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate

func BasicAuth(realm string, credentials map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			username, password, ok := req.BasicAuth()
			if !ok {
				unauthorized(w, realm)
				return
			}

			// Verify
			validPassword, found := credentials[username]
			if !found || subtle.ConstantTimeCompare([]byte(password), []byte(validPassword)) != 1 {
				unauthorized(w, realm)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// responds with 401 Unauthorized 
func unauthorized(w http.ResponseWriter, realm string) {
	// Set WWW-Authenticate header
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized) // 401
}