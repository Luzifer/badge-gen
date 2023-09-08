package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const liberapayCacheDuration = 60 * time.Minute

func init() {
	registerServiceHandler("liberapay", liberapayServiceHandler{})
}

type liberapayPublicProfile struct {
	Giving *struct {
		Amount   float64 `json:"amount,string"`
		Currency string  `json:"currency"`
	} `json:"giving"`
	Receiving *struct {
		Amount   float64 `json:"amount,string"`
		Currency string  `json:"currency"`
	} `json:"receiving"`
}

type liberapayServiceHandler struct{}

func (liberapayServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "LiberaPay Amount Receiving",
			DemoPath:    "/liberapay/liberapay/receiving",
			Arguments:   []string{"<user>", "receiving"},
		},
		{
			ServiceName: "LiberaPay Amount Giving",
			DemoPath:    "/liberapay/Nutomic/giving",
			Arguments:   []string{"<user>", "giving"},
		},
	}
}

func (liberapayServiceHandler) IsEnabled() bool { return true }

func (liberapayServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		err = errors.New("you need to provide user and payment direction")
		return title, text, color, err
	}

	if !str.StringInSlice(params[1], []string{"receiving", "giving"}) {
		err = fmt.Errorf("%q is an invalid payment direction", params[1])
		return title, text, color, err
	}

	title = params[1]
	color = colorNameBrightGreen

	cacheKey := strings.Join([]string{params[0], params[1]}, ":")
	text, err = cacheStore.Get("liberapay", cacheKey)

	if err != nil {
		req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://liberapay.com/%s/public.json", params[0]), nil)

		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return title, text, color, errors.Wrap(err, "executing request")
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logrus.WithError(err).Error("closing response body (leaked fd)")
			}
		}()

		r := liberapayPublicProfile{}
		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return title, text, color, errors.Wrap(err, "decoding JSON response")
		}

		switch params[1] {
		case "receiving":
			if r.Receiving == nil {
				text = "hidden"
			} else {
				text = fmt.Sprintf("%.2f %s", r.Receiving.Amount, r.Receiving.Currency)
			}
		case "giving":
			if r.Giving == nil {
				text = "hidden"
			} else {
				text = fmt.Sprintf("%.2f %s", r.Giving.Amount, r.Giving.Currency)
			}
		}

		logErr(cacheStore.Set("liberapay", cacheKey, text, liberapayCacheDuration), "writing liberapay result to cache")
	}

	return title, text, color, nil
}
