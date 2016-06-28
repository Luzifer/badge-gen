package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

func init() {
	registerServiceHandler("travis", travisServiceHandler{})
}

type travisServiceHandler struct{}

func (t travisServiceHandler) GetDocumentation() serviceHandlerDocumentation {
	return serviceHandlerDocumentation{
		ServiceName: "Travis-CI",
		DemoPath:    "/travis/Luzifer/password",
		Arguments:   []string{"<user>", "<repo>", "[branch]"},
	}
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

	var resp *http.Response
	resp, err = ctxhttp.Get(ctx, http.DefaultClient, "https://api.travis-ci.org/"+path)
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

	title = "travis"
	text = r.Branch.State
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
