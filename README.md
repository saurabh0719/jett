<div align="center">
    <img src="https://github.com/saurabh0719/jett/blob/assets/assets/jett_new.png" width="50%">
	<br>
    <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/saurabh0719/jett?style=for-the-badge"> <img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/saurabh0719/jett?color=FFD500&style=for-the-badge">
</div>
<hr>

Jett is a lightweight micro-framework for building Go HTTP services. Built on top of [HttpRouter](https://github.com/julienschmidt/httprouter). 

Jett strives to be simple, without unnecessary abstractions, rather letting the router and methods from `net/http` shine. This allows Jett to be extremely flexible right out of the box. 

The core framework is less than 300 loc but is designed to be easily extendable with middleware.

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
	jett.JSON(w, "Hello World", 200)
}

// Middleware
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Middleware\n")
		next.ServeHTTP(w, req)
	})
}
```

Install - 

```sh
$ go get github.com/saurabh0719/jett
```

<span id="keyfeatures"></span>

### Key Features :
* Build REST APIs quickly with minimal abstraction! 

* Add middleware at any level - Root, Subrouter or in a specific route!
* Built-in development server with support for graceful shutdown and shutdown functions.
* Highly Flexible & easily customisable with middleware.
* Helpful Response writers for JSON, XML and Plain Text.
* Extremely lightweight. Built on top of HttpRouter.

<hr>

<span id="contents"></span>

### Table of Contents :
* [Using Middleware](#middleware)
* [Subrouter](#subrouter)
* [Register Routes](#routes)
* [Path & Query parameters](#params)
* [Development Server](#devserver)
* [Response Writers](#writers)
* [Contribute](#contributors)

<hr>

<span id="middleware"></span>

### Using Middleware 

```go
func (r *Router) Use(middleware ...func(http.Handler) http.Handler)
```

Middleware can be added at the at a Router level (root, subrouter) ... 

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

### Subrouter 

The `Subrouter` function returns a new `Router` instance.

Example - 

```go 
func main() {

	r := jett.New()

	r.Use(Logger)

	r.GET("/", Home)

	sr := r.Subrouter("/about")
	sr.Use(Recover)
	sr.GET("/", About)

	r.Run(":8000")
}
```

<hr> 

<span id="routes"></span>

### Register routes 

```go 
// These functions optionally accept their own unique middleware for their handlers

func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PUT(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) POST(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PATCH(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) DELETE(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) OPTIONS(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)
```

Optionally, You can directly call the `Handle` function that accepts an `http.Handler`

```go
func (r *Router) Handle(method, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler)
```

[Go back to the table of contents](#contents)

<hr>

<span id="params"></span>

### Path & Query parameters 

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

	r.Run(":8000")
}

func Person(w http.ResponseWriter, req *http.Request) {
	params := jett.PathParams(req)
	
    // do something and prepare resp

	jett.JSON(w, resp, http.StatusOK)
}
```

[Go back to the table of contents](#contents)

<hr>

<span id="devserver"></span>

### Development Server

Jett comes with a built-in development server that handles graceful shutdown. You can optionally specify multiple cleanup functions to be called on shutdown. 

#### Run without context - 

```go
func (r *Router) Run(address string, onShutdownFns ...func())
```

```go
func (r *Router) RunTLS(address, certFile, keyFile string, onShutdownFns ...func())
```

#### Run with context - 

Useful when you need to pass a top-level context. Shutdown process will begin when the parent context cancels.

```go
func (r *Router) RunWithContext(ctx context.Context, address string, onShutdownFns ...func())
```

```go
func (r *Router) RunTLSWithContext(ctx context.Context, address, certFile, keyFile string, onShutdownFns ...func())
```

Example - 

`server.go` 

```go 

package main

import (
	"fmt"
	"github.com/saurabh0719/jett"
	"net/http"
	"time"
)

func main() {

	r := jett.New()

	r.GET("/", Home)

	// automatically trigger shutdown after 10s
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r.RunWithContext(ctx, ":8000", shutdownOne, shutdownTwo)
}

func Home(w http.ResponseWriter, req *http.Request) {
	jett.TEXT(w, "Hello World!", 200)
}

// Shutdown functions called during graceful shutdown
func shutdownOne() {
	time.Sleep(1 * time.Second)
	fmt.Println("shutdown function 1 complete!")
}

func shutdownTwo() {
	time.Sleep(1 * time.Second)
	fmt.Println("shutdown function 2 complete!")
}

```

```sh
$ go run server.go
```

Please note that this Server is for development only. A production server should ideally specify timeouts inside http.Server. Any contributions to build upon this is welcome.

[Go back to the table of contents](#contents)

<hr>

<span id="writers"></span>

### Response Writers

Optional helpers for formatting the output 

```go 
// JSON output
func JSON(w http.ResponseWriter, data interface{}, status int)

// Plain Text output
func TEXT(w http.ResponseWriter, data string, status int)

// XML output
func XML(w http.ResponseWriter, data interface{}, status int)
```
<hr>

<span id="contributors"></span>

### Contribute

Author and maintainer - [Saurabh Pujari](https://github.com/saurabh0719)

Logo design - [Akhil Anil](https://twitter.com/adakidpv)

Actively looking for Contributors to further improve upon this project. If you have any interesting ideas
or feature suggestions, don't hesitate to open an issue! First thing on the To-do list is to add a few essential middleware in a separate `middleware/` package!

[Go back to the table of contents](#contents)

<hr>
