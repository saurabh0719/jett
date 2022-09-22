package middleware

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saurabh0719/jett"
)

func handler(w http.ResponseWriter, req *http.Request) {
	reqID := req.Context().Value("requestID")
	jett.JSON(w, reqID, 200)
}

func TestMiddlewareRequestIDWithCustomHeaderStrKey(t *testing.T) {
	var headerKey = "X-Request-ID"
	var headerValue = "12345"
	r := jett.New()

	r.Use(RequestIDWithCustomHeaderStrKey(headerKey))

	r.GET("/", handler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set(headerKey, headerValue)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	// Convert []byte to string
	var requestId string
	json.Unmarshal(body, &requestId)

	if requestId != headerValue {
		t.Fatalf("middleware.RequestID -> Expected : %s, Output : %s", headerValue, requestId)
	}
}
