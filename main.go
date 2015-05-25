package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"
)

const (
	xSpacing = 8
)

func main() {
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":3000"
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/badge", generateBadge).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(port, nil)
}

func generateBadge(res http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	text := r.URL.Query().Get("text")

	if title == "" || text == "" {
		http.Error(res, "You must specify parameters 'title' and 'text'.", http.StatusInternalServerError)
		return
	}

	badge := createBadge(title, text)

	res.Header().Add("Content-Type", "image/svg+xml")
	res.Header().Add("Cache-Control", "public, max-age=31536000")
	res.Write(badge)
}

func createBadge(title, text string) []byte {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	titleW, _ := calculateTextWidth(title)
	textW, _ := calculateTextWidth(text)

	width := titleW + textW + 4*xSpacing

	t, _ := Asset("assets/badgeTemplate.tpl")
	tpl, _ := template.New("svg").Parse(string(t))

	tpl.Execute(bufw, map[string]interface{}{
		"Width":       width,
		"TitleWidth":  titleW + 2*xSpacing,
		"Title":       title,
		"Text":        text,
		"TitleAnchor": titleW/2 + xSpacing,
		"TextAnchor":  titleW + textW/2 + 3*xSpacing,
	})

	bufw.Flush()
	return buf.Bytes()
}
