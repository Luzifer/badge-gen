package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

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

func (a aurServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
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

func (a aurServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		return title, text, color, errors.New("No service-command / parameters were given")
	}

	switch params[0] {
	case "license":
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

		cacheStore.Set("aur_license", title, text, 10*time.Minute)
	}

	return "license", text, "blue", nil
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

		cacheStore.Set("aur_version", title, text, 10*time.Minute)
	}

	return title, text, "blue", nil
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
			text = text + " (outdated)"
		}

		cacheStore.Set("aur_updated", title, text, 10*time.Minute)
	}

	color = "blue"
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

		cacheStore.Set("aur_votes", title, text, 10*time.Minute)
	}

	return title, text, "brightgreen", nil
}

func (a aurServiceHandler) fetchAURInfo(ctx context.Context, pkg string) (*aurInfoResult, error) {
	params := url.Values{
		"v":    []string{"5"},
		"type": []string{"info"},
		"arg":  []string{pkg},
	}
	url := "https://aur.archlinux.org/rpc/?" + params.Encode()

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch AUR info")
	}
	defer resp.Body.Close()

	out := &aurInfoResult{}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return nil, errors.Wrap(err, "Failed to parse AUR info")
	}

	if out.Resultcount == 0 {
		return nil, errors.New("No package was found")
	}

	return out, nil
}
