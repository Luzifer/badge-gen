package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

var (
	githubRemainingLimit prometheus.Gauge
)

func init() {
	registerServiceHandler("github", githubServiceHandler{})
	githubRemainingLimit = prometheus.MustRegisterOrGet(prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "github_remaining_limit",
		Help: "Remaining requests in GitHub rate limit until next reset",
	})).(prometheus.Gauge)
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name      string `json:"name"`
	Downloads int64  `json:"download_count"`
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
		{
			ServiceName: "GitHub downloads by repo",
			DemoPath:    "/github/downloads/atom/atom",
			Arguments:   []string{"downloads", "<user>", "<repo>"},
		},
		{
			ServiceName: "GitHub downloads by release",
			DemoPath:    "/github/downloads/atom/atom/latest",
			Arguments:   []string{"downloads", "<user>", "<repo>", "<tag or \"latest\">"},
		},
		{
			ServiceName: "GitHub downloads by release and asset",
			DemoPath:    "/github/downloads/atom/atom/v1.8.0/atom-amd64.deb",
			Arguments:   []string{"downloads", "<user>", "<repo>", "<tag or \"latest\">", "<asset>"},
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
	case "downloads":
		title, text, color, err = g.handleDownloads(ctx, params[1:])
	default:
		err = errors.New("An unknown service command was called")
	}

	return
}

func (g githubServiceHandler) handleDownloads(ctx context.Context, params []string) (title, text, color string, err error) {
	switch len(params) {
	case 2:
		title, text, color, err = g.handleRepoDownloads(ctx, params)
	case 3:
		params = append(params, "total")
		fallthrough
	case 4:
		title, text, color, err = g.handleReleaseDownloads(ctx, params)
	default:
		err = errors.New("Unsupported number of arguments")
	}
	return
}

func (g githubServiceHandler) handleReleaseDownloads(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases", "tags", params[2]}, "/")
	if params[2] == "latest" {
		path = strings.Join([]string{"repos", params[0], params[1], "releases", params[2]}, "/")
	}

	text, err = cacheStore.Get("github_release_downloads", path)

	if err != nil {
		r := githubRelease{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return
		}

		var sum int64

		for _, rel := range r.Assets {
			if params[3] == "total" || rel.Name == params[3] {
				sum = sum + rel.Downloads
			}
		}

		text = metricFormat(sum)
		cacheStore.Set("github_release_downloads", path, text, 10*time.Minute)
	}

	title = "downloads"
	color = "brightgreen"
	return
}

func (g githubServiceHandler) handleRepoDownloads(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases"}, "/")

	text, err = cacheStore.Get("github_repo_downloads", path)

	if err != nil {
		r := []githubRelease{}

		// TODO: This does not respect pagination!
		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return
		}

		var sum int64

		for _, rel := range r {
			for _, rea := range rel.Assets {
				sum = sum + rea.Downloads
			}
		}

		text = metricFormat(sum)
		cacheStore.Set("github_repo_downloads", path, text, 10*time.Minute)
	}

	title = "downloads"
	color = "brightgreen"
	return
}

func (g githubServiceHandler) handleLatestRelease(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases", "latest"}, "/")

	text, err = cacheStore.Get("github_latest_release", path)

	if err != nil {
		r := githubRelease{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return
		}

		text = r.TagName
		if text == "" {
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

	if rlr := resp.Header.Get("X-RateLimit-Remaining"); rlr != "" {
		v, err := strconv.ParseInt(rlr, 10, 64)
		if err == nil {
			githubRemainingLimit.Set(float64(v))
		}
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
