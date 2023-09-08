package main

import (
	"errors"

	"golang.org/x/net/context"
)

func init() {
	registerServiceHandler("static", staticServiceHandler{})
}

type staticServiceHandler struct{}

func (staticServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{{
		ServiceName: "Static Badge",
		DemoPath:    "/static/API/Documentation/4c1",
		Arguments:   []string{"<title>", "<text>", "[color]"},
	}}
}

func (staticServiceHandler) IsEnabled() bool { return true }

func (staticServiceHandler) Handle(_ context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		err = errors.New("you need to provide title and text")
		return title, text, color, err
	}

	if len(params) < 3 { //nolint:gomnd
		params = append(params, defaultColor)
	}

	title = params[0]
	text = params[1]
	color = params[2]
	return title, text, color, err
}
