package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/str"
	"golang.org/x/net/context"
)

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

func (s liberapayServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
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

func (s liberapayServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide user and payment direction")
		return
	}

	if !str.StringInSlice(params[1], []string{"receiving", "giving"}) {
		err = fmt.Errorf("%q is an invalid payment direction", params[1])
		return
	}

	title = params[1]
	color = "ffee16"

	cacheKey := strings.Join([]string{params[0], params[1]}, ":")
	text, err = cacheStore.Get("liberapay", cacheKey)

	if err != nil {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://liberapay.com/%s/public.json", params[0]), nil)

		var resp *http.Response
		resp, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		r := liberapayPublicProfile{}
		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return
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

		cacheStore.Set("liberapay", cacheKey, text, 60*time.Minute)
	}

	return
}
