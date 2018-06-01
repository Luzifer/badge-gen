package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func init() {
	registerServiceHandler("travis", travisServiceHandler{})
}

type travisServiceHandler struct{}

func (t travisServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{{
		ServiceName: "Travis-CI",
		DemoPath:    "/travis/Luzifer/password",
		Arguments:   []string{"<user>", "<repo>", "[branch]"},
	}}
}

func (t travisServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide user and repo")
		return
	}

	if len(params) < 3 {
		params = append(params, "master")
	}

	path := strings.Join([]string{"repos", params[0], params[1], "branches", params[2]}, "/")

	var state string
	state, err = cacheStore.Get("travis", path)

	if err != nil {
		var resp *http.Response
		req, _ := http.NewRequest("GET", "https://api.travis-ci.org/"+path, nil)
		resp, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		r := struct {
			File   string `json:"file"`
			Branch struct {
				State string `json:"state"`
			} `json:"branch"`
		}{}

		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return
		}
		state = r.Branch.State
		cacheStore.Set("travis", path, state, 5*time.Minute)
	}

	title = "travis"
	text = state
	if text == "" {
		text = "unknown"
	}
	color = map[string]string{
		"unknown":  "9f9f9f",
		"passed":   "4c1",
		"failed":   "e05d44",
		"canceled": "9f9f9f",
	}[text]

	return
}
