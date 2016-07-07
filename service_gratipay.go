package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

func init() {
	registerServiceHandler("gratipay", gratipayServiceHandler{})
}

type gratipayServiceHandler struct{}

func (s gratipayServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "Gratipay user",
			DemoPath:    "/gratipay/user/js_bin",
			Arguments:   []string{"user", "<username>"},
		},
		{
			ServiceName: "Gratipay team",
			DemoPath:    "/gratipay/team/Gratipay",
			Arguments:   []string{"team", "<teamname>"},
		},
	}
}

func (s gratipayServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide type and user/teamname")
		return
	}

	if params[0] == "user" {
		params[1] = "~" + params[1]
	}

	path := strings.Join([]string{params[1], "public.json"}, "/")
	text, err = cacheStore.Get("gratipay", path)

	if err != nil {
		apiURL := "https://gratipay.com/" + path

		r := struct {
			Receiving float64 `json:"receiving"`
			Taking    float64 `json:"taking,string"`
		}{}

		var resp *http.Response
		resp, err = ctxhttp.Get(ctx, http.DefaultClient, apiURL)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return
		}

		sum := r.Taking + r.Receiving
		color = "brightgreen"

		if sum == 0.0 {
			color = "red"
		} else if sum < 10 {
			color = "yellow"
		} else if sum < 100 {
			color = "green"
		}

		text = fmt.Sprintf("$%.2f / week::%s", sum, color)
		cacheStore.Set("gratipay", path, text, 10*time.Minute)
	}

	tmp := strings.Split(text, "::")
	text = tmp[0]
	color = tmp[1]

	title = "gratipay"
	return
}
