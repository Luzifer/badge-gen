package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateBadge(t *testing.T) {
	badge := string(createBadge("API", "Documentation", "4c1"))

	if !strings.Contains(badge, ">API</text>") {
		t.Error("Did not found node with text 'API'")
	}

	if !strings.Contains(badge, "<path fill=\"#4c1\"") {
		t.Error("Did not find color coding for path")
	}

	if !strings.Contains(badge, ">Documentation</text>") {
		t.Error("Did not found node with text 'Documentation'")
	}
}

func TestHttpResponseMissingParameters(t *testing.T) {
	resp := httptest.NewRecorder()

	uri := "/v1/badge"

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	generateMux().ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Response code should be %d, is %d", http.StatusInternalServerError, resp.Code)
		}

		if string(p) != "You must specify parameters 'title' and 'text'.\n" {
			t.Error("Response message did not match test")
		}
	}

}

func TestHttpResponseWithoutColor(t *testing.T) {
	resp := httptest.NewRecorder()

	uri := "/v1/badge?"
	params := url.Values{
		"title": []string{"API"},
		"text":  []string{"Documentation"},
	}

	req, err := http.NewRequest("GET", uri+params.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}

	generateMux().ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		if resp.Code != http.StatusOK {
			t.Errorf("Response code should be %d, is %d", http.StatusInternalServerError, resp.Code)
		}

		if resp.Header().Get("Content-Type") != "image/svg+xml" {
			t.Errorf("Response had wrong Content-Type: %s", resp.Header().Get("Content-Type"))
		}

		// Check whether there is a SVG in the response, format checks are in other checks
		if !strings.Contains(string(p), "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"133\" height=\"20\">") {
			t.Error("Response message did not match test")
		}

		if !strings.Contains(string(p), "#4c1") {
			t.Error("Default color was not set")
		}
	}

}

func TestHttpResponseWithColor(t *testing.T) {
	resp := httptest.NewRecorder()

	uri := "/v1/badge?"
	params := url.Values{
		"title": []string{"API"},
		"text":  []string{"Documentation"},
		"color": []string{"572"},
	}

	req, err := http.NewRequest("GET", uri+params.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}

	generateMux().ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		if resp.Code != http.StatusOK {
			t.Errorf("Response code should be %d, is %d", http.StatusInternalServerError, resp.Code)
		}

		if resp.Header().Get("Content-Type") != "image/svg+xml" {
			t.Errorf("Response had wrong Content-Type: %s", resp.Header().Get("Content-Type"))
		}

		// Check whether there is a SVG in the response, format checks are in other checks
		if !strings.Contains(string(p), "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"133\" height=\"20\">") {
			t.Error("Response message did not match test")
		}

		if strings.Contains(string(p), "#4c1") {
			t.Error("Default color is present with color given")
		}

		if !strings.Contains(string(p), "#572") {
			t.Error("Given color is not present in SVG")
		}
	}

}
