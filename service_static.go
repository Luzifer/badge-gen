package main

import (
	"errors"

	"golang.org/x/net/context"
)

func init() {
	registerServiceHandler("static", staticServiceHandler{})
}

type staticServiceHandler struct{}

func (s staticServiceHandler) GetDocumentation() serviceHandlerDocumentation {
	return serviceHandlerDocumentation{
		ServiceName: "Static Badge",
		DemoPath:    "/static/API/Documentation/4c1",
		Arguments:   []string{"<title>", "<text>", "[color]"},
	}
}

func (s staticServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 {
		err = errors.New("You need to provide title and text")
		return
	}

	if len(params) < 3 {
		params = append(params, defaultColor)
	}

	title = params[0]
	text = params[1]
	color = params[2]
	return
}
