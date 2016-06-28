package main

import "errors"

func init() {
	registerServiceHandler("static", func(params []string) (title, text, color string, err error) {
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
	})
}
