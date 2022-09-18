<div align="center">
    <img src="assets/jett.png">
	![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/saurabh0719/jett?style=for-the-badge)
</div>
<hr>

Jett is a lightweight micro-framework for quickly building Go HTTP services. Built on top of HttpRouter. 

Jett strives to be simple, without unnecessary abstractions, rather letting the router and methods from `net/http` shine. This allows Jett to be extremely flexible right out of the box. 

The core framework is less than `300 loc` but is designed to be easily extandable with middleware.


```go
package main

import (
	"fmt"
	"net/http"
	"github.com/saurabh0719/jett"
)

func main() {

	r := jett.New()

	r.Use(Logger)

	r.GET("/", Home)
	
	r.Run(":8000")
}

func Home(w http.ResponseWriter, req *http.Request) {
	jett.JSONResponse(w, "Hello World", 200)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Middleware\n")
		next.ServeHTTP(w, req)
	})
}
```

<hr>

<span id="contents"></span>

## Table of Contents :
* [Key Features](#keyfeatures)
* [Using Middleware](#middleware)
* [Subrouter](#subrouter)
* [Development Server](#devserver)
* [Register Routes](#routes)
* [Path & Query parameters](#params)
* [Response Writers](#writers)
* [Contributors](#contributors)

<hr>

<span id="keyfeatures"></span>

## Key Features :
* Build Robust APIs with minimal abstraction! 

* Add middleware at the root level or to a specific route.
* Built-in development server with support for graceful shutdown with timeout and shutdown functions.
* Highly Flexible & easily customisable with middleware.
* Helpful Response writers for JSON, XML and Plain Text.
* Extremely lightweight. Built on top of HttpRouter.

<hr>

<span id="middleware"></span>

## Using Middleware 

```go
func (r *Router) Use(middleware ...func(http.Handler) http.Handler)
```

Middleware can be added at the root... 

```go
func main() {

	r := jett.New()

	r.GET("/", Home, Logger, Recover)
	
	r.Run(":8000")
}
```
OR on each individual route

```go
func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)
```

Example - 

```go
func main() {

	r := jett.New()

	r.GET("/", Home, Logger, Recover)
	
	r.Run(":8000")
}
```

Compatible with any Middleware of the type `func(http.Handler) http.Handler`

[Go back to the table of contents](#contents)

<hr>

<span id="subrouter"></span>

## Subrouter 

Subrouters cannot have their own Middleware. But you can add specific middleware to each route of a subrouter. 

Example - 

```go 
func main() {

	r := jett.New()

	r.Use(Logger)

	r.GET("/", Home)
	
	r.Run(":8000")

	sr := r.Subrouter("/about")
	sr.GET("/", About)

	h.RunServer(":8000", 5)
}
```

<hr> 

<span id="devserver"></span>

## Development Server

```go
func (r *Router) RunServer(address string, timeout int, onShutdownFns ...func())
```

`RunServer` creates a server and allows for graceful shutdown. You can specify a `timeout` (seconds) before the server closes the context. You can also pass multiple cleanup functions (`onShutdownFns ...func()`) to run on shutdown.

Apart from `RunServer` - Jett also has helper functions  `func (r *Router) Run(address string)` and `func (r *Router) RunTLS(address, certFile, keyFile string)`.

[Go back to the table of contents](#contents)

<hr>

<span id="routes"></span>

## Register routes 

```go 
// These functions optionally accept their own unique middleware for their handlers

func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PUT(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) POST(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PATCH(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) DELETE(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) OPTIONS(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)
```

[Go back to the table of contents](#contents)

<hr>

<span id="params"></span>

## Path & Query parameters 

Path parameters -
```go
// Helper function to extract path params from request Context()
// as a map[string]string for easy access
func PathParams(req *http.Request) map[string]string
```

Query parameters - 
```go
// Helper function to extract query params as a map[string][]string
// Eg. /?one=true,false&two=true
// {"two" : ["true"], "one": ["true, "false"]}
func QueryParams(req *http.Request) map[string][]string
```

Example - 
```go
func main() {

	r := jett.New()

	r.GET("/person/:id", Person)

	h.RunServer(":8000", 5)
}

func About(w http.ResponseWriter, req *http.Request) {
	params := jett.PathParams(req)
	
    // do something 

	JSONResponse(w, resp, http.StatusOK)
}
```

[Go back to the table of contents](#contents)

<hr>

<span id="writers"></span>

## Response Writers

Optional helpers for formatting the output 

```go 
// JSON output
func JSONResponse(w http.ResponseWriter, data interface{}, status int)

// Plain Text output
func PlainResponse(w http.ResponseWriter, data string, status int)

// XML output
func XMLResponse(w http.ResponseWriter, data interface{}, status int)
```
<hr>

<span id="contributors"></span>
Author and maintainer - [Saurabh Pujari](https://github.com/saurabh0719)

Logo design - [Akhil Anil](https://twitter.com/adakidpv)

[Go back to the table of contents](#contents)

<hr>
