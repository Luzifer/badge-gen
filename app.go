package main

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"text/template"

	"github.com/Luzifer/rconfig"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
)

const (
	xSpacing     = 8
	defaultColor = "4c1"
)

var (
	cfg = struct {
		Port   int64  `env:"PORT"`
		Listen string `flag:"listen" default:":3000" description:"Port/IP to listen on"`
	}{}
	serviceHandlers = map[string]serviceHandler{}
	version         = "dev"
)

type serviceHandlerDocumentation struct {
	ServiceName string
	DemoPath    string
	Arguments   []string
	Register    string
}

func (s serviceHandlerDocumentation) DocFormat() string {
	return "/" + s.Register + "/" + strings.Join(s.Arguments, "/")
}

type serviceHandlerDocumentationList []serviceHandlerDocumentation

func (s serviceHandlerDocumentationList) Len() int { return len(s) }
func (s serviceHandlerDocumentationList) Less(i, j int) bool {
	return s[i].ServiceName < s[j].ServiceName
}
func (s serviceHandlerDocumentationList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type serviceHandler interface {
	GetDocumentation() serviceHandlerDocumentation
	Handle(params []string) (title, text, color string, err error)
}

func registerServiceHandler(service string, f serviceHandler) error {
	if _, ok := serviceHandlers[service]; ok {
		return errors.New("Duplicate service handler")
	}
	serviceHandlers[service] = f
	return nil
}

func main() {
	rconfig.Parse(&cfg)
	if cfg.Port != 0 {
		cfg.Listen = fmt.Sprintf(":%d", cfg.Port)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/badge", generateBadge).Methods("GET")
	r.HandleFunc("/{service}/{parameters:.*}", generateServiceBadge).Methods("GET")
	r.HandleFunc("/", handleDemoPage)

	http.ListenAndServe(cfg.Listen, r)
}

func generateServiceBadge(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	params := strings.Split(vars["parameters"], "/")

	handler, ok := serviceHandlers[service]
	if !ok {
		http.Error(res, "Service not found: "+service, http.StatusNotFound)
		return
	}

	title, text, color, err := handler.Handle(params)
	if err != nil {
		http.Error(res, "Error while executing service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderBadgeToResponse(res, r, title, text, color)
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
		color = defaultColor
	}

	renderBadgeToResponse(res, r, title, text, color)
}

func renderBadgeToResponse(res http.ResponseWriter, r *http.Request, title, text, color string) {
	badge, eTag := createBadge(title, text, color)

	res.Header().Add("Cache-Control", "no-cache")
	res.Header().Add("ETag", eTag)

	if r.Header.Get("If-None-Match") == eTag {
		res.WriteHeader(http.StatusNotModified)
		return
	}

	res.Header().Add("Content-Type", "image/svg+xml")

	m := minify.New()
	m.AddFunc("image/svg+xml", svg.Minify)

	badge, _ = m.Bytes("image/svg+xml", badge)

	res.Write(badge)
}

func createBadge(title, text, color string) ([]byte, string) {
	var buf bytes.Buffer

	titleW, _ := calculateTextWidth(title)
	textW, _ := calculateTextWidth(text)

	width := titleW + textW + 4*xSpacing

	t, _ := Asset("assets/badgeTemplate.tpl")
	tpl, _ := template.New("svg").Parse(string(t))

	tpl.Execute(&buf, map[string]interface{}{
		"Width":       width,
		"TitleWidth":  titleW + 2*xSpacing,
		"Title":       title,
		"Text":        text,
		"TitleAnchor": titleW/2 + xSpacing,
		"TextAnchor":  titleW + textW/2 + 3*xSpacing,
		"Color":       color,
	})

	return buf.Bytes(), generateETag(buf.Bytes())
}

func generateETag(in []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(in))
}

func handleDemoPage(res http.ResponseWriter, r *http.Request) {
	t, _ := Asset("assets/demoPage.tpl.html")
	tpl, _ := template.New("demoPage").Parse(string(t))

	examples := serviceHandlerDocumentationList{}

	for register, handler := range serviceHandlers {
		tmp := handler.GetDocumentation()
		tmp.Register = register
		examples = append(examples, tmp)
	}

	sort.Sort(examples)

	tpl.Execute(res, map[string]interface{}{
		"Examples": examples,
		"Version":  version,
	})
}
