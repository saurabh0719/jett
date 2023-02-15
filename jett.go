// Jett is a lightweight micro-framework for building Go HTTP services.
//
// Jett builds a layer on top of HttpRouter to enable subrouting
// and flexible addition of middleware at any level - root, subrouter or a specific route!
//
// Built for Go 1.7 & above.
//
// Example :
// 	package main
//
// 	import (

// 		"net/http"
// 		"github.com/saurabh0719/jett"
// 		"github.com/saurabh0719/jett/middleware"
// 	)
//
// 	func main() {
//
// 		r := jett.New()
//
// 		r.Use(middleware.RequestID, middleware.Logger)
//
// 		r.GET("/", Home)
//
// 		r.Run(":8000")
// 	}
//
// 	func Home(w http.ResponseWriter, req *http.Request) {
// 		jett.JSON(w, "Hello World", 200)
// 	}
//
//
// Jett strives to be simple and easy to use with minimal abstractions.
// The core framework is less than 300 loc but is designed to be extendable with middleware.
// Comes packaged with a development server equipped for graceful shutdown
// and a few essential middleware.
//
// Read https://github.com/saurabh0719/jett#readme for further details.
//
// LICENSE
//
// BSD 3-Clause License.
// Copyright (c) 2022, Saurabh Pujari.
// All rights reserved.
//
package jett

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Jett package version
const (
	Version = "0.3.0"
	website = "https://www.github.com/saurabh0719/jett"
	banner  = `     ____.         __     __    
    |    |  ____ _/  |_ _/  |_  
    |    |_/ __ \\   __\\   __\ 
/\__|    |\  ___/ |  |   |  |   
\________| \____ >|__|   |__|  
	`
)

var httpMethods = [...]string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
}

// Jett's Router is built on top of julienschmidt's httprouter.
// https://github.com/julienschmidt/httprouter
type Router struct {
	// httprouter struct
	router *httprouter.Router

	// middleware stack -> List of middleware associated with the router
	middleware []func(http.Handler) http.Handler

	// pathPrefix -> Contains total path of that router,
	// which is then prefixed with every subrouter.
	// default - '/' (root)
	pathPrefix string
}

// Create a new instance of the Jett's Router
func New() *Router {

	// new instance of httprouter
	r := httprouter.New()

	// Recommended to set to false
	// See README.md - https://github.com/julienschmidt/httprouter/
	r.HandleMethodNotAllowed = false

	return &Router{
		router: r,
		// Root path prefix
		pathPrefix: "/",
	}
}

/* -------------------------- Router Methods  ------------------------- */

// Add a middlware to the Router's middlware stack.
// To use built-in essential middleware,
//	 import "github.com/saurabh0719/jett/middleware"
// Read https://github.com/saurabh0719/jett#middleware for further details.
func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

// Create a new subrouter.
// The subrouter automatically gets assigned the middleware from the parent router
func (r *Router) Subrouter(path string) *Router {

	sr := &Router{
		router:     r.router,
		middleware: r.middleware,
		pathPrefix: r.getFullPath(path),
	}

	return sr
}

// Assigns a HandlerFunc as http NotFound handler
func (r *Router) NotFound(handlerFn http.HandlerFunc) {
	r.router.NotFound = http.HandlerFunc(handlerFn)
}

// creates an http.Handler for the router + middleware stack
func (r *Router) Handler() http.Handler {
	var handler http.Handler = r.router
	return handler
}

// Implement http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := r.Handler()
	handler.ServeHTTP(w, req)
}

// Middleware returns a slice ([]func(http.Handler) http.Handler) of the middleware stack for the router
func (r *Router) Middleware() []func(http.Handler) http.Handler {
	return r.middleware
}

// Serve Static files from a directory.
// From github.com/julienschmidt/httprouter -> router.go :
//
//  ServeFiles serves files from the given file system root.
//  The path must end with "/*filepath", files are then served from the local
//  path /defined/root/dir/*filepath.
//
//  For example if root is "/etc" and *filepath is "passwd", the local file
//  "/etc/passwd" would be served.
//
//  Internally a http.FileServer is used, therefore http.NotFound is used instead
//  of the Router's NotFound handler.
//
// 	To use the operating system's file system implementation,
//  	use http.Dir:
//     		router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.router.ServeFiles(path, root)
}

// Retrieves full path of the current handler from root
func (r *Router) getFullPath(subPath string) string {
	fullPath := r.pathPrefix + subPath
	// Removes duplicate/multiple slash(es)
	// pure canonical form
	return httprouter.CleanPath(fullPath)
}

/* -------------------------- REGISTER HTTP METHOD HANDLERS ------------------------- */

// Register the path and method to the given handler. Also applies the middleware to the Handler
func (r *Router) Handle(method, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) {

	// full path from root
	fullPath := r.getFullPath(path)

	// apply the middleware passed to the Handle method
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	// apple rest of the middleware stack from the Router
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	// insert into httprouter
	r.router.Handler(method, fullPath, handler)
}

// Assigns a HandlerFunc to the GET method for the given path. Route-specific middleware can be added as well.
func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodGet, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the HEAD method for the given path. Route-specific middleware can be added as well.
func (r *Router) HEAD(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodHead, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the OPTIONS method for the given path. Route-specific middleware can be added as well.
func (r *Router) OPTIONS(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodOptions, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the POST method for the given path. Route-specific middleware can be added as well.
func (r *Router) POST(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodPost, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the PUT method for the given path. Route-specific middleware can be added as well.
func (r *Router) PUT(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodPut, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the PATCH method for the given path. Route-specific middleware can be added as well.
func (r *Router) PATCH(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodPatch, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the DELETE method for the given path. Route-specific middleware can be added as well.
func (r *Router) DELETE(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle(http.MethodDelete, path, http.HandlerFunc(handlerFn), middleware...)
}

// Assigns a HandlerFunc to the GET, HEAD, OPTIONS, POST, PUT, PATCH & DELETE method for the given path.
// It DOES NOT actually match any random arbitrary method.
func (r *Router) Any(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	for _, method := range httpMethods {
		r.Handle(method, path, http.HandlerFunc(handlerFn), middleware...)
	}
}

/* -------------------------- GET PARAMS  ------------------------- */

// Helper function to extract URL params from request Context()
// as a map[string]string for easy access.
func URLParams(req *http.Request) map[string]string {

	var routerParams httprouter.Params
	routerParams = httprouter.ParamsFromContext(req.Context())

	var params = make(map[string]string)
	for _, item := range routerParams {
		params[item.Key] = item.Value
	}

	return params
}

// Helper function to extract query params as a map[string][]string
//
// Eg - /?one=true,false&two=true
//
// Result - {"two" : ["true"], "one": ["true, "false"]}
func QueryParams(req *http.Request) map[string][]string {
	return req.URL.Query()
}

/* -------------------------- DEVELOPMENT SERVER & Run Fns------------------------- */

//
// Jett's development server that handles graceful shutdown.
// - ctx -> coordinates shutdown with a top level context
// - onShutdownFns -> Cleanup functions to run during shutdown
//
// Please note that this Server is for development only.
// A production server should ideally specify timeouts inside http.Server
//
func (r *Router) runServer(ctx context.Context, address, certFile, keyFile string, onShutdownFns ...func()) {

	// Check if server needs to run with TLS protocol
	isTLS := true
	if certFile == "" && keyFile == "" {
		isTLS = false
	}

	// New http server
	server := &http.Server{
		Addr:    address,
		Handler: r,
	}

	// Notify stopServer channel with any of the below mentioned Signals
	stopServer := make(chan os.Signal, 1)
	signal.Notify(stopServer, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run Server
	go func() {
		if isTLS {
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Error: %s\n", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Error: %s\n", err)
			}
		}
	}()

	fmt.Println(banner)
	fmt.Println(website)

	fmt.Printf("Running Jett Server v%s, address -> %s\n\n", Version, address)

	// Stop the server on signal notif or when parent ctx cancels
	select {
	case <-stopServer:
	case <-ctx.Done():
	}

	fmt.Printf("\n")
	fmt.Println("-> Shutting down the server...")
	defer fmt.Println("-> Server exited successfully.")

	// context.Background() gives us an empty context
	// set timeout to avoid keeping zombie conns alive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Defer the running of shutdown functions
	defer func() {
		totalFns := len(onShutdownFns)
		if totalFns > 0 {
			fmt.Println("-> Running shutdown functions...")
		}

		// Call each shutdown function one by one
		for i, j := totalFns-1, 1; i >= 0; i, j = i-1, j+1 {
			fmt.Println("-> ", j, " of ", totalFns)
			onShutdownFns[i]()
		}

		// Stop receiving signals
		signal.Stop(stopServer)
		// Cancel context after timeout
		cancel()
	}()

	// Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("-> Server Shutdown Failed:%+v", err)
	}

}

//
// The following functions wrap around runServer to abstract certain functionality
// that may not suit your usecase.
//
// You can choose to run the server normally (ListenAndServe) with or with TLS and with/without a context.Context
// (in which case context.TODO() is set)
//

// development server that handles graceful shutdown.
// onShutdownFns -> Cleanup functions to run during shutdown
func (r *Router) Run(address string, onShutdownFns ...func()) {
	r.runServer(context.TODO(), address, "", "", onShutdownFns...)
}

// development server that handles graceful shutdown.
// ctx -> coordinates shutdown with a top level context
func (r *Router) RunWithContext(ctx context.Context, address string, onShutdownFns ...func()) {
	r.runServer(ctx, address, "", "", onShutdownFns...)
}

// development server that runs with TLS and handles graceful shutdown.
// onShutdownFns -> Cleanup functions to run during shutdown
func (r *Router) RunTLS(address, certFile, keyFile string, onShutdownFns ...func()) {
	r.runServer(context.TODO(), address, certFile, keyFile, onShutdownFns...)
}

// development server that runs with TLS and handles graceful shutdown.
// ctx -> coordinates shutdown with a top level context
func (r *Router) RunTLSWithContext(ctx context.Context, address, certFile, keyFile string, onShutdownFns ...func()) {
	r.runServer(ctx, address, certFile, keyFile, onShutdownFns...)
}

/* -------------------------- RESPONSE RENDERERS ------------------------ */

//
// Optional helper functions for standard JSON, XML or plain text responses.
// Enforces the need to explicitly declare an http status code.
// Also ensures the correct Content-Type header is set to avoid client rendering issues.
//

// JSON renderer.
// Sets the status code and the Content-Type header to application/json
func JSON(w http.ResponseWriter, data interface{}, status int) {
	// prepare JSON response
	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Print("Internal Server Error - JSON Response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// Plain Text renderer.
// Sets the status code and the Content-Type header to text/plain
func Text(w http.ResponseWriter, data string, status int) {
	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")

	// Write plain text response
	_, err := fmt.Fprintf(w, data)

	if err != nil {
		log.Print("Internal Server Error - Plain Text Response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// XML renderer.
// Sets the Content-Type header to application/xml
func XML(w http.ResponseWriter, data interface{}, status int) {
	// prepare XML response
	xmlData, err := xml.Marshal(data)

	if err != nil {
		log.Print("Internal Server Error - XML Response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/xml")
	w.Write(xmlData)
}

// HTML template renderer -
// Sets the Content-Type header to text/html.
// Can render nested html files. Files need to ne sent in order of parent -> children
func HTML(w http.ResponseWriter, data interface{}, htmlFiles ...string) {

	// Parse all the html files passed
	t, err := template.ParseFiles(htmlFiles...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// New buffer to store html
	htmlBuffer := new(bytes.Buffer)

	// pass data (or nil) for the template
	if err := t.Execute(htmlBuffer, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html")

	// Write plain text response
	_, err = fmt.Fprintf(w, htmlBuffer.String())
	if err != nil {
		log.Print("Internal Server Error - HTML Template Response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
