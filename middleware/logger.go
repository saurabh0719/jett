package middleware 

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

// Wraps http.ResponseWriter to allow us to store Status Code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}
  
func wrapWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
	}
}
  
func (rw *responseWriter) Status() int {
	return rw.status
}
  
// Implement WriteHeader for registering status code 
func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
	  return
	}
  
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
  
	return
}

// A basic logger for Jett
// Logs 
// 	- RequestID (if available from RequestID middleware)
// 	- Method and Path 
// 	- status code of response
// 	- Duration of the request-response cycle 
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		
		// Get unique requestID from request Context
		requestID := GetRequestID(req.Context())

		// START
		start := ""
		if requestID != "" {
			start = "START RequestID: " + requestID
		} else {
			start = "START RequestID: <nil>"
		}
		log.Print(start + " - " + req.Method + " " + req.URL.String())

		// register start time
		t1 := time.Now()

		// Wrap http.ResponseWriter
		wrapped := wrapWriter(w)

		// Call downstream handlers
		next.ServeHTTP(wrapped, req)

		// register end time
		t2 := time.Now()

		// END
		end := ""
		if requestID != "" {
			end = "  END RequestID: " + requestID
		} else {
			end = "  END RequestID: <nil>"
		}

		// Prepare duration log 
		duration := ""
		d := t2.Sub(t1)
		duration = "Duration: "  + d.String()

		// Prepare final log with Status code
		status := wrapped.Status()
		if status > 99 && status < 600 {
			log.Printf(end + " - " + "Status: " + strconv.Itoa(status) + ", " + duration + "\n")
		} else {
			log.Printf(end + " - " + duration + "\n")
		}
		
	})
}