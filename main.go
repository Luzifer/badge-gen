package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
)

const (
	xSpacing = 8
)

func main() {
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":3000"
	}

	http.Handle("/", generateMux())
	http.ListenAndServe(port, nil)
}

func generateMux() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/v1/badge", generateBadge).Methods("GET")

	return r
}

func generateBadge(res http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	text := r.URL.Query().Get("text")
	color := r.URL.Query().Get("color")

	if title == "" || text == "" {
		http.Error(res, "You must specify parameters 'title' and 'text'.", http.StatusInternalServerError)
		return
	}

	if color == "" {
		color = "4c1"
	}

	badge := createBadge(title, text, color)

	res.Header().Add("Content-Type", "image/svg+xml")
	res.Header().Add("Cache-Control", "public, max-age=31536000")

	m := minify.New()
	m.AddFunc("image/svg+xml", svg.Minify)

	badge, _ = minify.Bytes(m, "image/svg+xml", badge)

	res.Write(badge)
}

func createBadge(title, text, color string) []byte {
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
		"Color":       color,
	})

	bufw.Flush()
	return buf.Bytes()
}
