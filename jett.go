/*
BSD 3-Clause License

Copyright (c) 2022, Saurabh Pujari
All rights reserved.
*/

package jett

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Jett package details
const (
	Version = "0.1.1"
	website = "https://www.github.com/saurabh0719/jett"
	banner  = `     ____.         __     __    
    |    |  ____ _/  |_ _/  |_  
    |    |_/ __ \\   __\\   __\ 
/\__|    |\  ___/ |  |   |  |   
\________| \____ >|__|   |__|  
	`
)

/*

Jett's Router is built upon @julienschmidt's httprouter

- middleware []func(http.Handler) http.Handler -> List of middleware associated with the (sub)router
- pathPrefix -> Contains total path of that router, which is then prefixed with every subrouter.

*/

type Router struct {
	// httprouter struct
	router *httprouter.Router

	// middleware stack
	middleware []func(http.Handler) http.Handler

	// default - '/' (root)
	pathPrefix string
}

// Create a new instance of the Router
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

// Add a middlware to the Router's middlware stack
func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

// Create a new subrouter
func (r *Router) Subrouter(path string) *Router {

	sr := &Router{
		router:     r.router,
		middleware: r.middleware,
		pathPrefix: r.getFullPath(path),
	}

	return sr
}

// Retrieves full path of the current handler from root
func (r *Router) getFullPath(subPath string) string {
	fullPath := r.pathPrefix + subPath
	// Removes duplicate/multiple slash(es)
	// pure canonical form
	return httprouter.CleanPath(fullPath)
}

// Assigns a function as http NotFound handler
func (r *Router) NotFound(handlerFn http.HandlerFunc) {
	r.router.NotFound = http.HandlerFunc(handlerFn)
}

// Serve Static files from a directory
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.router.ServeFiles(path, root)
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

/* -------------------------- REGISTER HTTP METHOD HANDLERS ------------------------- */

// Register a the given handler
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

// These functions optionally accept their own unique middleware for their handlers

func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("GET", path, http.HandlerFunc(handlerFn), middleware...)
}

func (r *Router) PUT(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("PUT", path, http.HandlerFunc(handlerFn), middleware...)
}

func (r *Router) POST(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("POST", path, http.HandlerFunc(handlerFn), middleware...)
}

func (r *Router) PATCH(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("PATCH", path, http.HandlerFunc(handlerFn), middleware...)
}

func (r *Router) DELETE(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("DELETE", path, http.HandlerFunc(handlerFn), middleware...)
}

func (r *Router) OPTIONS(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	r.Handle("OPTIONS", path, http.HandlerFunc(handlerFn), middleware...)
}

/* -------------------------- GET PARAMS  ------------------------- */

// Helper function to extract path params from request Context()
// as a map[string]string for easy access
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
// Eg. /?one=true,false&two=true
// {"two" : ["true"], "one": ["true, "false"]}
func QueryParams(req *http.Request) map[string][]string {
	return req.URL.Query()
}

/* -------------------------- DEVELOPMENT SERVER & Run Fns------------------------- */

/*

Jett's development server that handles graceful shutdown.
- ctx -> coordinates shutdown with a top level context
- onShutdownFns -> Cleanup functions to run during shutdown

Please note that this Server is for development only.
A production server should ideally specify timeouts inside http.Server

*/

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

/*

The following functions wrap around runServer to abstract certain functionality
that may not suit your usecase.

You can choose to run the server normally or with TLS and with/without a context.Context
(in which case context.TODO() is set)

*/

func (r *Router) Run(address string, onShutdownFns ...func()) {
	r.runServer(context.TODO(), address, "", "", onShutdownFns...)
}

func (r *Router) RunWithContext(ctx context.Context, address string, onShutdownFns ...func()) {
	r.runServer(ctx, address, "", "", onShutdownFns...)
}

func (r *Router) RunTLS(address, certFile, keyFile string, onShutdownFns ...func()) {
	r.runServer(context.TODO(), address, certFile, keyFile, onShutdownFns...)
}

func (r *Router) RunTLSWithContext(ctx context.Context, address, certFile, keyFile string, onShutdownFns ...func()) {
	r.runServer(ctx, address, certFile, keyFile, onShutdownFns...)
}

/* -------------------------- RESPONSE WRITERS  ------------------------- */

/*

Optional helper functions for standard JSON, XML or plain text responses

- Enforces the need to explicitly declare an http status code
- Also ensures the correct Content-Type header is set to avoid client rendering issues

*/

// JSON output
func JSON(w http.ResponseWriter, data interface{}, status int) {
	// prepare JSON response
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Print("Internal Server Error - JSONResponse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// Plain Text output
func TEXT(w http.ResponseWriter, data string, status int) {
	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")

	// Write plain text response
	_, err := fmt.Fprintf(w, data)
	if err != nil {
		log.Print("Internal Server Error - PlainResponse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// XML output
func XML(w http.ResponseWriter, data interface{}, status int) {
	// prepare XML response
	xmlData, err := xml.Marshal(data)
	if err != nil {
		log.Print("Internal Server Error - XMLResponse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set status and Content-Type
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/xml")
	w.Write(xmlData)
}

// Coming soon - helpers for templates/static files
