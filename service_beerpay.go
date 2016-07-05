package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
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

func (s beerpayServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide user and project")
		return
	}

	var resp *http.Response

	apiURL := fmt.Sprintf("https://beerpay.io/api/v1/%s/projects/%s", params[0], params[1])
	resp, err = ctxhttp.Get(ctx, http.DefaultClient, apiURL)
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

	title = "beerpay"
	text = fmt.Sprintf("$%d", r.TotalAmount)
	color = "red"
	return
}
