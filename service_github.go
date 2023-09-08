package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const githubCacheDuration = 10 * time.Minute

func init() {
	registerServiceHandler("github", githubServiceHandler{})
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name      string `json:"name"`
	Downloads int64  `json:"download_count"`
}

type githubRepo struct {
	StargazersCount int64 `json:"stargazers_count"`
}

type githubServiceHandler struct{}

func (githubServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
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
		{
			ServiceName: "Github stars by repository",
			DemoPath:    "/github/stars/atom/atom",
			Arguments:   []string{"stars", "<user>", "<repo>"},
		},
	}
}

func (githubServiceHandler) IsEnabled() bool { return true }

func (g githubServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		err = errors.New("no service-command / parameters were given")
		return title, text, color, err
	}

	switch params[0] {
	case "license":
		title, text, color, err = g.handleLicense(ctx, params[1:])
	case "latest-tag":
		title, text, color, err = g.handleLatestTag(ctx, params[1:])
	case "latest-release":
		title, text, color, err = g.handleLatestRelease(ctx, params[1:])
	case "downloads": //nolint:goconst
		title, text, color, err = g.handleDownloads(ctx, params[1:])
	case "stars":
		title, text, color, err = g.handleStargazers(ctx, params[1:])
	default:
		err = errors.New("an unknown service command was called")
	}

	return title, text, color, err
}

func (g githubServiceHandler) handleStargazers(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1]}, "/")

	text, err = cacheStore.Get("github_repo_stargazers", path)

	if err != nil {
		r := githubRepo{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return title, text, color, err
		}

		text = metricFormat(r.StargazersCount)
		logErr(cacheStore.Set("github_repo_stargazers", path, text, githubCacheDuration), "writing Github repo stargazers to cache")
	}

	title = "stars"
	color = colorNameBrightGreen
	return title, text, color, err
}

func (g githubServiceHandler) handleDownloads(ctx context.Context, params []string) (title, text, color string, err error) {
	switch len(params) {
	case 2: //nolint:gomnd
		title, text, color, err = g.handleRepoDownloads(ctx, params)
	case 3: //nolint:gomnd
		params = append(params, "total")
		fallthrough
	case 4: //nolint:gomnd
		title, text, color, err = g.handleReleaseDownloads(ctx, params)
	default:
		err = errors.New("Unsupported number of arguments")
	}
	return title, text, color, err
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
			return title, text, color, err
		}

		var sum int64

		for _, rel := range r.Assets {
			if params[3] == "total" || rel.Name == params[3] {
				sum += rel.Downloads
			}
		}

		text = metricFormat(sum)
		logErr(cacheStore.Set("github_release_downloads", path, text, githubCacheDuration), "writing Github release downloads to cache")
	}

	title = "downloads"
	color = colorNameBrightGreen
	return title, text, color, err
}

func (g githubServiceHandler) handleRepoDownloads(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases"}, "/")

	text, err = cacheStore.Get("github_repo_downloads", path)

	if err != nil {
		r := []githubRelease{}

		// NOTE: This does not respect pagination!
		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return title, text, color, err
		}

		var sum int64

		for _, rel := range r {
			for _, rea := range rel.Assets {
				sum += rea.Downloads
			}
		}

		text = metricFormat(sum)
		logErr(cacheStore.Set("github_repo_downloads", path, text, githubCacheDuration), "writing Github repo downloads to cache")
	}

	title = "downloads"
	color = colorNameBrightGreen
	return title, text, color, err
}

func (g githubServiceHandler) handleLatestRelease(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "releases", "latest"}, "/")

	text, err = cacheStore.Get("github_latest_release", path)

	if err != nil {
		r := githubRelease{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return title, text, color, err
		}

		text = r.TagName
		if text == "" {
			text = "None" //nolint:goconst
		}
		logErr(cacheStore.Set("github_latest_release", path, text, githubCacheDuration), "writing Github last release to cache")
	}

	title = "release"
	color = colorNameBlue

	if regexp.MustCompile(`^v?0\.`).MatchString(text) {
		color = "orange"
	}

	return title, text, color, err
}

func (g githubServiceHandler) handleLatestTag(ctx context.Context, params []string) (title, text, color string, err error) {
	path := strings.Join([]string{"repos", params[0], params[1], "tags"}, "/")

	text, err = cacheStore.Get("github_latest_tag", path)

	if err != nil {
		r := []struct {
			Name string `json:"name"`
		}{}

		if err = g.fetchAPI(ctx, path, nil, &r); err != nil {
			return title, text, color, err
		}

		if len(r) > 0 {
			text = r[0].Name
		} else {
			text = "None"
		}
		logErr(cacheStore.Set("github_latest_tag", path, text, githubCacheDuration), "writing Github last tag to cache")
	}

	title = "tag"
	color = colorNameBlue

	if regexp.MustCompile(`^v?0\.`).MatchString(text) {
		color = "orange"
	}

	return title, text, color, err
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
			return title, text, color, err
		}

		text = r.License.Name
		logErr(cacheStore.Set("github_license", path, text, githubCacheDuration), "writing Github license to cache")
	}

	title = "license"
	color = "007ec6"

	if text == "" {
		text = "None"
	}

	return title, text, color, err
}

func (githubServiceHandler) fetchAPI(ctx context.Context, path string, headers map[string]string, out interface{}) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/"+path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// #configStore github.username - string - Username for Github auth to increase API requests
	// #configStore github.personal_token - string - Token for Github auth to increase API requests
	if configStore.Str("github.personal_token") != "" {
		req.SetBasicAuth(configStore.Str("github.username"), configStore.Str("github.personal_token"))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "executing HTTP request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing request body (leaked fd)")
		}
	}()

	return errors.Wrap(
		json.NewDecoder(resp.Body).Decode(out),
		"decoding JSON response",
	)
}
