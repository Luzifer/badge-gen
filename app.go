package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/badge-gen/cache"
	"github.com/Luzifer/go_helpers/v2/accessLogger"
	"github.com/Luzifer/rconfig/v2"
)

const (
	badgeGenerationTimeout = 1500 * time.Millisecond
	xSpacing               = 8
	defaultColor           = "4c1"
)

const (
	colorNameBlue        = "blue"
	colorNameBrightGreen = "brightgreen"
	colorNameGray        = "gray"
	colorNameGreen       = "green"
	colorNameLightGray   = "lightgray"
	colorNameOrange      = "orange"
	colorNameRed         = "red"
	colorNameYellow      = "yellow"
	colorNameYellowGreen = "yellowgreen"
)

var (
	cfg = struct {
		LogLevel    string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		Port        int64  `env:"PORT"`
		Listen      string `flag:"listen" default:":3000" description:"Port/IP to listen on"`
		Cache       string `flag:"cache" default:"mem://" description:"Where to cache query results from thirdparty APIs"`
		ConfStorage string `flag:"config" default:"config.yaml" description:"Configuration store"`
	}{}

	serviceHandlers = map[string]serviceHandler{}
	version         = "dev"

	colorList = map[string]string{
		colorNameBlue:        "007ec6",
		colorNameBrightGreen: "4c1",
		colorNameGray:        "555",
		colorNameGreen:       "97CA00",
		colorNameLightGray:   "9f9f9f",
		colorNameOrange:      "fe7d37",
		colorNameRed:         "e05d44",
		colorNameYellow:      "dfb317",
		colorNameYellowGreen: "a4a61d",
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

func registerServiceHandler(service string, f serviceHandler) {
	if _, ok := serviceHandlers[service]; ok {
		panic("duplicate service handler")
	}

	serviceHandlers[service] = f
}

func initApp() error {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		return errors.Wrap(err, "parsing commandline options")
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(l)

	if cfg.Port != 0 {
		cfg.Listen = fmt.Sprintf(":%d", cfg.Port)
	}

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	logrus.Infof("badge-gen %s started...", version)

	cacheStore, err = cache.GetCacheByURI(cfg.Cache)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open cache")
	}

	f, err := os.Open(cfg.ConfStorage)
	switch {
	case err == nil:
		yamlDecoder := yaml.NewDecoder(f)
		yamlDecoder.SetStrict(true)
		if err = yamlDecoder.Decode(&configStore); err != nil {
			logrus.WithError(err).Fatal("Unable to parse config")
		}
		logrus.Printf("Loaded %d value pairs into configuration store", len(configStore))

		if err = f.Close(); err != nil {
			logrus.WithError(err).Error("closing config file (leaked fd)")
		}

	case os.IsNotExist(err):
		// Do nothing

	default:
		logrus.WithError(err).Fatal("Unable to open config")
	}

	r := mux.NewRouter().UseEncodedPath()
	r.HandleFunc("/v1/badge", generateBadge).Methods("GET")
	r.HandleFunc("/{service}/{parameters:.*}", generateServiceBadge).Methods("GET")
	r.HandleFunc("/", handleDemoPage)

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("HTTP server exited unexpectedly")
	}
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

	ctx, cancel := context.WithTimeout(r.Context(), badgeGenerationTimeout)
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
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s::::%s::::%s", title, text, color))))
	storedTag, _ := cacheStore.Get("eTag", cacheKey)

	res.Header().Add("Cache-Control", "no-cache")

	if storedTag != "" && r.Header.Get("If-None-Match") == storedTag {
		res.Header().Add("ETag", storedTag)
		res.WriteHeader(http.StatusNotModified)
		return
	}

	badge, eTag := createBadge(title, text, color)
	_ = cacheStore.Set("eTag", cacheKey, eTag, time.Hour)

	res.Header().Add("ETag", eTag)
	res.Header().Add("Content-Type", "image/svg+xml")

	m := minify.New()
	m.AddFunc("image/svg+xml", svg.Minify)

	badge, _ = m.Bytes("image/svg+xml", badge)

	if _, err := res.Write(badge); err != nil {
		logrus.WithError(err).Error("writing badge")
	}
}

func createBadge(title, text, color string) ([]byte, string) {
	var buf bytes.Buffer

	titleW, _ := calculateTextWidth(title)
	textW, _ := calculateTextWidth(text)

	width := titleW + textW + 4*xSpacing //nolint:gomnd

	t, _ := assets.ReadFile("assets/badgeTemplate.tpl")
	tpl, _ := template.New("svg").Parse(string(t))

	if c, ok := colorList[color]; ok {
		color = c
	}

	_ = tpl.Execute(&buf, map[string]any{
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
	return fmt.Sprintf("%x", sha256.Sum256(in))
}

func handleDemoPage(res http.ResponseWriter, _ *http.Request) {
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

	if err := tpl.Execute(res, map[string]interface{}{
		"Examples": examples,
		"Version":  version,
	}); err != nil {
		logrus.WithError(err).Error("rendering demo page")
	}
}

func logErr(err error, text string) {
	if err != nil {
		logrus.WithError(err).Error(text)
	}
}
