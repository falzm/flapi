package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/negroni"
)

type metricsMiddleware struct {
	ignore     *mux.Router
	reqLatency *prometheus.SummaryVec
}

func newMetricsMiddleware(service string, reqLatencyObjectives map[float64]float64, ignore *mux.Router) *metricsMiddleware {
	mw := metricsMiddleware{
		ignore: ignore,
	}

	if reqLatencyObjectives == nil {
		reqLatencyObjectives =
			map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			}
	}

	mw.reqLatency = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        "http_request_latency_seconds",
		Help:        "HTTP requests processing latency in seconds.",
		ConstLabels: prometheus.Labels{"service": service},
		Objectives:  reqLatencyObjectives,
		MaxAge:      1 * time.Minute,
		AgeBuckets:  1,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(mw.reqLatency)

	return &mw
}

func (mw *metricsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	var routeMatch mux.RouteMatch
	if mw.ignore != nil && mw.ignore.Match(r, &routeMatch) {
		return
	}

	res := rw.(negroni.ResponseWriter)

	mw.reqLatency.WithLabelValues(strconv.Itoa(res.Status()), r.Method, r.URL.Path).
		Observe(float64(time.Since(start).Nanoseconds()) / 1000000000)
}

func (m *metricsMiddleware) ServeMetrics() http.Handler {
	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
}
