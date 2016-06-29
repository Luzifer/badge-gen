package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

func init() {
	registerServiceHandler("github", githubServiceHandler{})
}

type githubServiceHandler struct{}

func (g githubServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "GitHub repo license",
			DemoPath:    "/github/license/Luzifer/badge-gen",
			Arguments:   []string{"license", "<user>", "<repo>"},
		},
	}
}

func (g githubServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("No service-command / parameters were given")
		return
	}

	switch params[0] {
	case "license":
		title, text, color, err = g.handleLicense(ctx, params[1:])
	default:
		err = errors.New("An unknown service command was called")
	}

	return
}

func (g githubServiceHandler) handleLicense(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "license"}, "/")

	text, err = cacheStore.Get("github_license", path)

	if err != nil {
		req, _ := http.NewRequest("GET", "https://api.github.com/"+path, nil)
		req.Header.Set("Accept", "application/vnd.github.drax-preview+json")

		var resp *http.Response
		resp, err = ctxhttp.Do(ctx, http.DefaultClient, req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		r := struct {
			License struct {
				Name string `json:"name"`
			} `json:"license"`
		}{}
		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return
		}

		text = r.License.Name
		cacheStore.Set("github_license", path, text, 10*time.Minute)
	}

	title = "license"
	color = "007ec6"

	if text == "" {
		text = "None"
	}

	return
}
