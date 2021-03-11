package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

func init() {
	registerServiceHandler("beerpay", beerpayServiceHandler{})
}

type beerpayServiceHandler struct{}

func (s beerpayServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{{
		ServiceName: "beerpay Total Amount",
		DemoPath:    "/beerpay/beerpay/beerpay.io",
		Arguments:   []string{"<user>", "<project>"},
	}}
}

func (beerpayServiceHandler) IsEnabled() bool { return true }

func (s beerpayServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide user and project")
		return
	}

	title = "beerpay"
	color = "red"

	cacheKey := fmt.Sprintf("%s::%s", params[0], params[1])
	text, err = cacheStore.Get("beerpay", cacheKey)

	if err != nil {
		var resp *http.Response

		apiURL := fmt.Sprintf("https://beerpay.io/api/v1/%s/projects/%s", params[0], params[1])
		req, _ := http.NewRequest("GET", apiURL, nil)
		resp, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		r := struct {
			TotalAmount int `json:"total_amount"`
		}{}

		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return
		}
		text = fmt.Sprintf("$%d", r.TotalAmount)
		cacheStore.Set("beerpay", cacheKey, text, 5*time.Minute)
	}

	return
}
