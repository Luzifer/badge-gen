package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const travisCacheDuration = 5 * time.Minute

func init() {
	registerServiceHandler("travis", travisServiceHandler{})
}

type travisServiceHandler struct{}

func (travisServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{{
		ServiceName: "Travis-CI",
		DemoPath:    "/travis/Luzifer/password",
		Arguments:   []string{"<user>", "<repo>", "[branch]"},
	}}
}

func (travisServiceHandler) IsEnabled() bool { return true }

func (travisServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		err = errors.New("you need to provide user and repo")
		return title, text, color, err
	}

	if len(params) < 3 { //nolint:gomnd
		params = append(params, "master")
	}

	path := strings.Join([]string{"repos", params[0], params[1], "branches", params[2]}, "/")

	var state string
	state, err = cacheStore.Get("travis", path)

	if err != nil {
		var resp *http.Response
		req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.travis-ci.org/"+path, nil)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return title, text, color, errors.Wrap(err, "executing request")
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logrus.WithError(err).Error("closing request body (leaked fd)")
			}
		}()

		r := struct {
			File   string `json:"file"`
			Branch struct {
				State string `json:"state"`
			} `json:"branch"`
		}{}

		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return title, text, color, errors.Wrap(err, "decoding JSON response")
		}
		state = r.Branch.State
		logErr(cacheStore.Set("travis", path, state, travisCacheDuration), "writing Travis status to cache")
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

	return title, text, color, nil
}
