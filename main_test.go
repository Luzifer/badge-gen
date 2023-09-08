package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Luzifer/badge-gen/cache"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func testGenerateMux() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/v1/badge", generateBadge).Methods("GET")
	m.HandleFunc("/{service}/{parameters:.*}", generateServiceBadge).Methods("GET")
	m.HandleFunc("/", handleDemoPage)
	return m
}

func TestMain(m *testing.M) {
	cacheStore = cache.NewInMemCache()
	os.Exit(m.Run())
}

func TestCreateBadge(t *testing.T) {
	badgeData, _ := createBadge("API", "Documentation", "4c1")
	badge := string(badgeData)

	assert.Contains(t, badge, ">API</text>")
	assert.Contains(t, badge, "<path fill=\"#4c1\"")
	assert.Contains(t, badge, ">Documentation</text>")
}

func TestHttpResponseMissingParameters(t *testing.T) {
	resp := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/v1/badge", nil)
	if err != nil {
		t.Fatal(err)
	}

	testGenerateMux().ServeHTTP(resp, req)
	if p, err := io.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, string(p), "You must specify parameters 'title' and 'text'.")
	}
}

func TestHttpResponseWithoutColor(t *testing.T) {
	resp := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/static/API/Documentation", nil)
	if err != nil {
		t.Fatal(err)
	}

	testGenerateMux().ServeHTTP(resp, req)
	if p, err := io.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "image/svg+xml", resp.Header().Get("Content-Type"))
		// Check whether there is a SVG in the response, format checks are in other checks
		assert.Contains(t, string(p), "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"133\" height=\"20\">")
		assert.Contains(t, string(p), "#4c1", "default color should be set")
	}
}

func TestHttpResponseWithColor(t *testing.T) {
	resp := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/static/API/Documentation/572", nil) //nolint:noctx // fine for an internal test
	if err != nil {
		t.Fatal(err)
	}

	testGenerateMux().ServeHTTP(resp, req)
	if p, err := io.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "image/svg+xml", resp.Header().Get("Content-Type"))
		// Check whether there is a SVG in the response, format checks are in other checks
		assert.Contains(t, string(p), "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"133\" height=\"20\">")
		assert.NotContains(t, string(p), "#4c1", "default color should not be set")
		assert.Contains(t, string(p), "#572", "given color should be set")
	}
}
