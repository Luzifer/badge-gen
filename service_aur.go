package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const aurCacheDuration = 10 * time.Minute

func init() {
	registerServiceHandler("aur", aurServiceHandler{})
}

type aurServiceHandler struct{}

type aurInfoResult struct {
	Version     int    `json:"version"`
	Type        string `json:"type"`
	Resultcount int    `json:"resultcount"`
	Results     []struct {
		ID             int      `json:"ID"`
		Name           string   `json:"Name"`
		PackageBaseID  int      `json:"PackageBaseID"`
		PackageBase    string   `json:"PackageBase"`
		Version        string   `json:"Version"`
		Description    string   `json:"Description"`
		URL            string   `json:"URL"`
		NumVotes       int      `json:"NumVotes"`
		Popularity     float64  `json:"Popularity"`
		OutOfDate      int      `json:"OutOfDate"`
		Maintainer     string   `json:"Maintainer"`
		FirstSubmitted int      `json:"FirstSubmitted"`
		LastModified   int      `json:"LastModified"`
		URLPath        string   `json:"URLPath"`
		Depends        []string `json:"Depends"`
		License        []string `json:"License"`
		Keywords       []string `json:"Keywords"`
		MakeDepends    []string `json:"MakeDepends,omitempty"`
	} `json:"results"`
}

func (aurServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "AUR package version",
			DemoPath:    "/aur/version/yay",
			Arguments:   []string{"version", "<package name>"},
		},
		{
			ServiceName: "AUR package votes",
			DemoPath:    "/aur/votes/yay",
			Arguments:   []string{"votes", "<package name>"},
		},
		{
			ServiceName: "AUR package license",
			DemoPath:    "/aur/license/yay",
			Arguments:   []string{"license", "<package name>"},
		},
		{
			ServiceName: "AUR package last update",
			DemoPath:    "/aur/updated/yay",
			Arguments:   []string{"updated", "<package name>"},
		},
	}
}

func (aurServiceHandler) IsEnabled() bool { return true }

func (a aurServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		return title, text, color, errors.New("No service-command / parameters were given")
	}

	switch params[0] {
	case "license": //nolint:goconst
		return a.handleAURLicense(ctx, params[1:])
	case "updated":
		return a.handleAURUpdated(ctx, params[1:])
	case "version":
		return a.handleAURVersion(ctx, params[1:])
	case "votes":
		return a.handleAURVotes(ctx, params[1:])
	default:
		return title, text, color, errors.New("An unknown service command was called")
	}
}

func (a aurServiceHandler) handleAURLicense(ctx context.Context, params []string) (title, text, color string, err error) {
	title = params[0]
	text, err = cacheStore.Get("aur_license", title)

	if err != nil {
		info, err := a.fetchAURInfo(ctx, params[0])
		if err != nil {
			return title, text, color, err
		}

		text = strings.Join(info.Results[0].License, ", ")

		logErr(cacheStore.Set("aur_license", title, text, aurCacheDuration), "writing AUR license to cache")
	}

	return "license", text, colorNameBlue, nil
}

func (a aurServiceHandler) handleAURVersion(ctx context.Context, params []string) (title, text, color string, err error) {
	title = params[0]
	text, err = cacheStore.Get("aur_version", title)

	if err != nil {
		info, err := a.fetchAURInfo(ctx, params[0])
		if err != nil {
			return title, text, color, err
		}

		text = info.Results[0].Version

		logErr(cacheStore.Set("aur_version", title, text, aurCacheDuration), "writing AUR version to cache")
	}

	return title, text, colorNameBlue, nil
}

func (a aurServiceHandler) handleAURUpdated(ctx context.Context, params []string) (title, text, color string, err error) {
	title = params[0]
	text, err = cacheStore.Get("aur_updated", title)

	if err != nil {
		info, err := a.fetchAURInfo(ctx, params[0])
		if err != nil {
			return title, text, color, err
		}

		update := time.Unix(int64(info.Results[0].LastModified), 0)
		text = update.Format("2006-01-02 15:04:05")

		if info.Results[0].OutOfDate > 0 {
			text += " (outdated)"
		}

		logErr(cacheStore.Set("aur_updated", title, text, aurCacheDuration), "writing AUR updated to cache")
	}

	color = colorNameBlue
	if strings.Contains(text, "outdated") {
		color = "red"
	}

	return "last updated", text, color, nil
}

func (a aurServiceHandler) handleAURVotes(ctx context.Context, params []string) (title, text, color string, err error) {
	title = params[0]
	text, err = cacheStore.Get("aur_votes", title)

	if err != nil {
		info, err := a.fetchAURInfo(ctx, params[0])
		if err != nil {
			return title, text, color, err
		}

		text = strconv.Itoa(info.Results[0].NumVotes) + " votes"

		logErr(cacheStore.Set("aur_votes", title, text, aurCacheDuration), "writing AUR votes to cache")
	}

	return title, text, colorNameBrightGreen, nil
}

func (aurServiceHandler) fetchAURInfo(ctx context.Context, pkg string) (*aurInfoResult, error) {
	params := url.Values{
		"v":    []string{"5"},
		"type": []string{"info"},
		"arg":  []string{pkg},
	}
	u := "https://aur.archlinux.org/rpc/?" + params.Encode()

	req, _ := http.NewRequest("GET", u, nil)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch AUR info")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	out := &aurInfoResult{}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return nil, errors.Wrap(err, "Failed to parse AUR info")
	}

	if out.Resultcount == 0 {
		return nil, errors.New("No package was found")
	}

	return out, nil
}
