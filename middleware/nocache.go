package middleware

// Adapted from Goji's nocache middleware
// Source: https://github.com/zenazn/goji/blob/master/web/middleware/nocache.go

import (
	"net/http"
	"time"
)

// NoCache is a simple piece of middleware that sets a number of HTTP headers to prevent
// a router (or subrouter) from being cached by an upstream proxy and/or client.
//
// As per http://wiki.nginx.org/HttpProxyModule - NoCache sets:
//      Expires: Thu, 01 Jan 1970 00:00:00 UTC
//      Cache-Control: no-cache, private, max-age=0
//      X-Accel-Expires: 0
//      Pragma: no-cache (for HTTP/1.0 proxies/clients)

func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Unix epoch time
		var epoch = time.Unix(0, 0).Format(time.RFC1123)

		// Taken from https://github.com/mytrile/nocache
		var noCacheHeaders = map[string]string{
			"Expires":         epoch,
			"Cache-Control":   "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
			"Pragma":          "no-cache",
			"X-Accel-Expires": "0",
		}

		var etagHeaders = []string{
			"ETag",
			"If-Modified-Since",
			"If-Match",
			"If-None-Match",
			"If-Range",
			"If-Unmodified-Since",
		}

		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if req.Header.Get(v) != "" {
				req.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		next.ServeHTTP(w, req)
		
	})
}