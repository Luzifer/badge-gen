package main

import "github.com/prometheus/client_golang/prometheus"

var (
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.SummaryVec
	responseSize    *prometheus.SummaryVec
)

func init() {
	initMetrics()
}

func initMetrics() {
	so := prometheus.SummaryOpts{
		Subsystem: "badge_gen",
	}

	reqCnt := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   so.Subsystem,
		Name:        "requests_total",
		Help:        "Total number of HTTP requests made.",
		ConstLabels: so.ConstLabels,
	}, []string{"handler", "method", "code"})

	so.Name = "response_size_bytes"
	so.Help = "The HTTP response sizes in bytes."
	resSz := prometheus.NewSummaryVec(so, []string{"handler"})

	so.Name = "request_duration_microseconds"
	so.Help = "The HTTP request latencies in microseconds."
	reqDur := prometheus.NewSummaryVec(so, []string{"handler"})

	requestCount = prometheus.MustRegisterOrGet(reqCnt).(*prometheus.CounterVec)
	requestDuration = prometheus.MustRegisterOrGet(reqDur).(*prometheus.SummaryVec)
	responseSize = prometheus.MustRegisterOrGet(resSz).(*prometheus.SummaryVec)
}
