package main

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/badge-gen/cache"
	"github.com/Luzifer/go_helpers/v2/accessLogger"
	"github.com/Luzifer/rconfig/v2"
)

const (
	xSpacing     = 8
	defaultColor = "4c1"
)

var (
	cfg = struct {
		Port        int64  `env:"PORT"`
		Listen      string `flag:"listen" default:":3000" description:"Port/IP to listen on"`
		Cache       string `flag:"cache" default:"mem://" description:"Where to cache query results from thirdparty APIs"`
		ConfStorage string `flag:"config" default:"config.yaml" description:"Configuration store"`
	}{}

	serviceHandlers = map[string]serviceHandler{}
	version         = "dev"

	colorList = map[string]string{
		"brightgreen": "4c1",
		"green":       "97CA00",
		"yellow":      "dfb317",
		"yellowgreen": "a4a61d",
		"orange":      "fe7d37",
		"red":         "e05d44",
		"blue":        "007ec6",
		"grey":        "555",
		"gray":        "555",
		"lightgrey":   "9f9f9f",
		"lightgray":   "9f9f9f",
	}

	cacheStore  cache.Cache
	configStore = configStorage{}
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
	return strings.ToLower(s[i].ServiceName) < strings.ToLower(s[j].ServiceName)
}
func (s serviceHandlerDocumentationList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type serviceHandler interface {
	GetDocumentation() serviceHandlerDocumentationList
	IsEnabled() bool
	Handle(ctx context.Context, params []string) (title, text, color string, err error)
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

	log.Infof("badge-gen %s started...", version)

	var err error
	cacheStore, err = cache.GetCacheByURI(cfg.Cache)
	if err != nil {
		log.WithError(err).Fatal("Unable to open cache")
	}

	f, err := os.Open(cfg.ConfStorage)
	switch {
	case err == nil:
		yamlDecoder := yaml.NewDecoder(f)
		yamlDecoder.SetStrict(true)
		if err = yamlDecoder.Decode(&configStore); err != nil {
			log.WithError(err).Fatal("Unable to parse config")
		}
		log.Printf("Loaded %d value pairs into configuration store", len(configStore))

		f.Close()

	case os.IsNotExist(err):
		// Do nothing

	default:
		log.WithError(err).Fatal("Unable to open config")
	}

	r := mux.NewRouter().UseEncodedPath()
	r.HandleFunc("/v1/badge", generateBadge).Methods("GET")
	r.HandleFunc("/{service}/{parameters:.*}", generateServiceBadge).Methods("GET")
	r.HandleFunc("/", handleDemoPage)

	http.ListenAndServe(cfg.Listen, r)
}

func generateServiceBadge(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]

	var err error
	params := strings.Split(vars["parameters"], "/")
	for i := range params {
		if params[i], err = url.QueryUnescape(params[i]); err != nil {
			http.Error(res, "Invalid escaping in URL", http.StatusBadRequest)
			return
		}
	}

	al := accessLogger.New(res)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	handler, ok := serviceHandlers[service]
	if !ok || !handler.IsEnabled() {
		http.Error(res, "Service not found: "+service, http.StatusNotFound)
		return
	}

	title, text, color, err := handler.Handle(ctx, params)
	if err != nil {
		http.Error(res, "Error while executing service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderBadgeToResponse(al, r, title, text, color)
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

	http.Redirect(res, r, fmt.Sprintf("/static/%s/%s/%s",
		url.QueryEscape(title),
		url.QueryEscape(text),
		url.QueryEscape(color),
	), http.StatusMovedPermanently)
}

func renderBadgeToResponse(res http.ResponseWriter, r *http.Request, title, text, color string) {
	cacheKey := fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%s::::%s::::%s", title, text, color))))
	storedTag, _ := cacheStore.Get("eTag", cacheKey)

	res.Header().Add("Cache-Control", "no-cache")

	if storedTag != "" && r.Header.Get("If-None-Match") == storedTag {
		res.Header().Add("ETag", storedTag)
		res.WriteHeader(http.StatusNotModified)
		return
	}

	badge, eTag := createBadge(title, text, color)
	cacheStore.Set("eTag", cacheKey, eTag, time.Hour)

	res.Header().Add("ETag", eTag)
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

	t, _ := assets.ReadFile("assets/badgeTemplate.tpl")
	tpl, _ := template.New("svg").Parse(string(t))

	if c, ok := colorList[color]; ok {
		color = c
	}

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
	t, _ := assets.ReadFile("assets/demoPage.tpl.html")
	tpl, _ := template.New("demoPage").Parse(string(t))

	examples := serviceHandlerDocumentationList{}

	for register, handler := range serviceHandlers {
		if !handler.IsEnabled() {
			continue
		}

		tmps := handler.GetDocumentation()
		for _, tmp := range tmps {
			tmp.Register = register
			examples = append(examples, tmp)
		}
	}

	sort.Sort(examples)

	tpl.Execute(res, map[string]interface{}{
		"Examples": examples,
		"Version":  version,
	})
}
