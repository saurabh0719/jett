<div align="center">
    <img src="https://github.com/saurabh0719/jett/blob/assets/assets/jett_new.png" width="50%">
	<br>
    <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/saurabh0719/jett?style=for-the-badge"> <img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/saurabh0719/jett?color=FFD500&style=for-the-badge">
</div>
<hr>

Jett is a lightweight micro-framework for building Go HTTP services. It builds a layer on top of [HttpRouter](https://github.com/julienschmidt/httprouter) to enable subrouting and flexible addition of middleware at any level - root, subrouter, a specific route.

Jett strives to be simple and easy to use with minimal abstractions. The core framework is less than 300 loc but is designed to be extendable with middleware. Comes packaged with a development server equipped for graceful shutdown and a few essential middleware.

<span id="keyfeatures"></span>

### Key Features :
* Build REST APIs quickly with minimal abstraction! 

* Add middleware at any level - Root, Subrouter or in a specific route!
* Built-in development server with support for graceful shutdown and shutdown functions.
* Highly flexible & easily customisable with middleware.
* Helpful response renderers for HTML, JSON, XML and Plain Text.
* Clean and simple code. Jett's core is less than 300 LOC.
* Extremely lightweight. Built on top of `net/http` & `HttpRouter`.

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/saurabh0719/jett"
	"github.com/saurabh0719/jett/middleware"
)

func main() {

	r := jett.New()

	r.Use(middleware.RequestID, middleware.Logger)

	r.GET("/", Home)
	
	r.Run(":8000")
}

func Home(w http.ResponseWriter, req *http.Request) {
	jett.JSON(w, "Hello World", 200)
}
```

Install - 

```sh
$ go get github.com/saurabh0719/jett
```

<hr>

<span id="contents"></span>

### Table of Contents :
* [Using middleware](#middleware)
* [Subrouter](#subrouter)
* [Register routes](#routes)
* [Static files](#static)
* [Path & Query parameters](#params)
* [Development server](#devserver)
* [Response renderers](#writers)
* [Contribute](#contributors)

<hr>

<span id="middleware"></span>

### Using Middleware 

Jett supports any Middleware of the type `func(http.Handler) http.Handler`. 

Some essential middleware are provided out of the box in `github.com/saurabh0719/jett/middleware` - 
- `RequestID` : Injects a request ID into the context of each
request

- `Logger` : Log request paths, methods, status code as well as execution duration 
- `Recoverer` : Recover and handle `panic` 
- `NoCache` : Sets a number of HTTP headers to prevent
a router (or subrouter) from being cached by an upstream proxy and/or client
- `BasicAuth` : Basic Auth middleware, [RFC 2617, Section 2](https://www.rfc-editor.org/rfc/rfc2617.html#section-2)

```go
func (r *Router) Use(middleware ...func(http.Handler) http.Handler)
```

Middleware can be added at the at a Router level (root, subrouter) ... 

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/saurabh0719/jett"
	"github.com/saurabh0719/jett/middleware"
)

func main() {

	r := jett.New()

	r.Use(middleware.RequestID, middleware.Logger)

	r.GET("/", Home)
	
	r.Run(":8000")
}
```
OR on each individual route

```go
func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)
```

To access a router's middleware stack - 
```go
// Middleware returns a slice of the middleware stack for the router
func (r *Router) Middleware() []func(http.Handler) http.Handler
```

Example - 

```go
func main() {

	r := jett.New()

	r.GET("/", Home, middleware.Logger, middleware.Recover)
	
	r.Run(":8000")
}
```

[Go back to the table of contents](#contents)

<hr>

<span id="subrouter"></span>

### Subrouter 

The `Subrouter` function returns a new `Router` instance.

Example - 

```go 
package main

import (
	"fmt"
	"net/http"
	"github.com/saurabh0719/jett"
	"github.com/saurabh0719/jett/middleware"
)
func main() {

	r := jett.New()

	r.Use(middleware.RequestID)

	r.GET("/", Home)

	sr := r.Subrouter("/about")
	sr.Use(middleware.Logger)
	sr.GET("/", About, middleware.NoCache)

	r.Run(":8000")
}

func Home(w http.ResponseWriter, req *http.Request) {
	jett.JSON(w, "Hello World", 200)
}

func About(w http.ResponseWriter, req *http.Request) {
	jett.TEXT(w, "About", 200)
}
```

<hr> 

<span id="routes"></span>

### Register routes 

```go 
// These functions optionally accept their own unique middleware for their handlers

func (r *Router) GET(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) HEAD(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) OPTIONS(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) POST(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PUT(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) PATCH(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

func (r *Router) DELETE(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)

// Any() creates routes for the methods mentioned above ^ - it DOES NOT actually match any random arbitrary method method
func (r *Router) Any(path string, handlerFn http.HandlerFunc, middleware ...func(http.Handler) http.Handler)
```

You can also directly call the `Handle` function that accepts an `http.Handler`

```go
func (r *Router) Handle(method, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler)
```

[Go back to the table of contents](#contents)

<hr>

<span id="static"></span>

### Serving static files

The `ServeFiles` is a wrapper around `httprouter.ServeFiles` to serve statice assets.

```go 
func (r *Router) ServeFiles(path string, root http.FileSystem)
```
From [HttpRouter](https://github.com/julienschmidt/httprouter) `router.go` 

> ServeFiles serves files from the given file system root.
> The path must end with "/*filepath", files are then served from the local
> path /defined/root/dir/*filepath.

> For example if root is "/etc" and *filepath is "passwd", the local file
> "/etc/passwd" would be served.
> Internally a http.FileServer is used, therefore http.NotFound is used instead of the Router's NotFound handler.

> To use the operating system's file system implementation,
> use http.Dir:
>     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))


Eg. `r.ServeFiles("/static/*filepath", http.Dir("static"))` 

[See a full example here](#example)

<hr> 

<span id="params"></span>

### Path & Query parameters 

Path parameters -
```go
// Helper function to extract path params from request Context()
// as a map[string]string for easy access
func URLParams(req *http.Request) map[string]string
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
	params := jett.URLParams(req)
	
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
	"context"
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

### Response renderers

Optional helpers for formatting the output. Content type is set automatically.

```go 
// JSON output - Content-Type - application/json
func JSON(w http.ResponseWriter, data interface{}, status int)

// Plain Text output - Content-Type - text/plain
func Text(w http.ResponseWriter, data string, status int)

// XML output - Content-Type - application/xml
func XML(w http.ResponseWriter, data interface{}, status int)
```

For html templates (status is set internally, default 200 OK else Server error)

```go 
// Content-Type - text/html
func HTML(w http.ResponseWriter, data interface{}, htmlFiles ...string) 
```

data can be `nil` or any struct that the template needs. You can also send multiple templates in order of parent -> child 

```go 
jett.HTML(w, nil, "layout.html", "index.html")
```

<span id="example"></span>

### A simple example - 

Directory Structure - 
```
example/
	- static/
		- styles.css
	- index.html
	- server.go
```

- `index.html`
```html

<!DOCTYPE html>
<html>
<head>
	<link rel="stylesheet" href="/static/styles.css">
</head>

<body>
	<h1>This is a heading</h1>
</body>
</html>
```

- `styles.css` 
```css
body {
    background-color: #FFD500;
}
```

- `server.go`
```go
package main

import (
	"fmt"
	"net/http"
	"github.com/saurabh0719/jett"
	"github.com/saurabh0719/jett/middleware"
)

func main() {

	r := jett.New()

	r.Use(middleware.RequestID, middleware.Logger)

	r.ServeFiles("/static/*filepath", http.Dir("static"))

	r.GET("/:name", Home)

	r.Run(":8000")
}

func Home(w http.ResponseWriter, req *http.Request) {
	params := jett.URLParams(req)
	p := Person{
		name: params["name"]
	}
	jett.HTML(w, p, "index.html")
}
```
<hr>

<span id="contributors"></span>

### Contribute

Author and maintainer - [Saurabh Pujari](https://github.com/saurabh0719)

Logo design - [Akhil Anil](https://twitter.com/adakidpv)

Actively looking for Contributors to further improve upon this project. If you have any interesting ideas
or feature suggestions, don't hesitate to open an issue! 

[Go back to the table of contents](#contents)

<hr>
