package middleware

// Ported from Chi's timeout middleware
// Source: https://github.com/go-chi/chi/blob/master/middleware/timeout.go

import (
	"context"
	"net/http"
	"time"
)

// Timeout is a middleware that cancels context after a given timeout and returns
// a 504 Gateway Timeout error to the client.
//
// It's required that you select the ctx.Done() channel to check for the signal
// if the context has reached its deadline and return, otherwise the timeout
// signal will be just ignored.
//
// Example route may look like:
//
//  r.GET("/task", func(w http.ResponseWriter, req *http.Request) {
// 	 ctx := r.Context()
// 	 processTime := time.Duration(rand.Intn(4)+1) * time.Second
//
// 	 select {
// 	 case <-ctx.Done():
// 	 	return
//
// 	 case <-time.After(processTime):
// 	 	 // The above channel simulates some hard work.
// 	 }
//
// 	 jett.JSON(w, "Done!", 200)
//
//  })
//
func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}