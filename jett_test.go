package jett

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPathParams(t *testing.T) {
	r := New()
	r.GET("/home/:param", Home)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/home/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	// Convert []byte to map
	var pathParams map[string]string
	json.Unmarshal(body, &pathParams)

	if pathParams["param"] != "hello" {
		t.Fatalf("PathParams -> Expected : [param] hello")
	}

}

func TestgetFullPath(t *testing.T) {
	r := New()

	expected := "/one"
	output := r.getFullPath("/one")

	if output != expected {
		t.Fatalf("getFullPath -> Expected : %s, Output : %s", expected, output)
	}

	r.pathPrefix = "/one/"
	expected = "/one/two"
	output = r.getFullPath("/two")

	if output != expected {
		t.Fatalf("getFullPath -> Expected : %s, Output : %s", expected, output)
	}
}

func TestSubrouter(t *testing.T) {

	r := New()

	r.GET("/", Home)

	sr := r.Subrouter("/about")
	sr.GET("/", About)

	if sr.pathPrefix != "/about" {
		t.Fatalf("Subrouter pathPrefix -> Expected : /about, Output : %s", sr.pathPrefix)
	}

}

func Home(w http.ResponseWriter, req *http.Request) {
	params := PathParams(req)
	JSONResponse(w, params, 200)
}

func About(w http.ResponseWriter, req *http.Request) {
	params := QueryParams(req)
	JSONResponse(w, params, 200)
}
