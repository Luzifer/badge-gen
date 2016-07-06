package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
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
		{
			ServiceName: "GitHub latest tag",
			DemoPath:    "/github/latest-tag/Luzifer/badge-gen",
			Arguments:   []string{"latest-tag", "<user>", "<repo>"},
		},
		{
			ServiceName: "GitHub latest release",
			DemoPath:    "/github/latest-release/lastpass/lastpass-cli",
			Arguments:   []string{"latest-release", "<user>", "<repo>"},
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
	case "latest-tag":
		title, text, color, err = g.handleLatestTag(ctx, params[1:])
	case "latest-release":
		title, text, color, err = g.handleLatestRelease(ctx, params[1:])
	default:
		err = errors.New("An unknown service command was called")
	}

	return
}

func (g githubServiceHandler) handleLatestRelease(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases"}, "/")

	text, err = cacheStore.Get("github_latest_release", path)

	if err != nil {
		r := []struct {
			TagName string `json:"tag_name"`
		}{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return
		}

		if len(r) > 0 {
			text = r[0].TagName
		} else {
			text = "None"
		}
		cacheStore.Set("github_latest_release", path, text, 10*time.Minute)
	}

	title = "release"
	color = "blue"

	if regexp.MustCompile(`^v?0\.`).MatchString(text) {
		color = "orange"
	}

	return
}

func (g githubServiceHandler) handleLatestTag(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "tags"}, "/")

	text, err = cacheStore.Get("github_latest_tag", path)

	if err != nil {
		r := []struct {
			Name string `json:"name"`
		}{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return
		}

		if len(r) > 0 {
			text = r[0].Name
		} else {
			text = "None"
		}
		cacheStore.Set("github_latest_tag", path, text, 10*time.Minute)
	}

	title = "tag"
	color = "blue"

	if regexp.MustCompile(`^v?0\.`).MatchString(text) {
		color = "orange"
	}

	return
}

func (g githubServiceHandler) handleLicense(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "license"}, "/")

	text, err = cacheStore.Get("github_license", path)

	if err != nil {
		r := struct {
			License struct {
				Name string `json:"name"`
			} `json:"license"`
		}{}

		headers := map[string]string{
			"Accept": "application/vnd.github.drax-preview+json",
		}
		if err = g.fetchAPI(ctx, path, headers, &r); err != nil {
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

func (g githubServiceHandler) fetchAPI(ctx context.Context, path string, headers map[string]string, out interface{}) error {
	req, _ := http.NewRequest("GET", "https://api.github.com/"+path, nil)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if configStore.Str("github.personal_token") != "" {
		req.SetBasicAuth(configStore.Str("github.username"), configStore.Str("github.personal_token"))
	}

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(out)
}
