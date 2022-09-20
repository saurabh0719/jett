package middleware

// Adapted from Goji's recoverer middleware
// Source: https://github.com/zenazn/goji/blob/master/web/middleware/recoverer.go

import (
	"log"
	"net/http"
	"runtime/debug"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Get unique requestID from request Context
		requestID := GetRequestID(req.Context())

		defer func() {
			err := recover()
			if err != nil {
				
				if requestID != "" {
					log.Println("RequestID: " + requestID)
				}

				log.Printf("Panic : %+v", err)
				debug.PrintStack()

				// Internal server error; No more writes to this Writer
				http.Error(w, http.StatusText(500), 500)

			}

		}()

		next.ServeHTTP(w, req)

	})
}