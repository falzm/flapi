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
	reqLatency *prometheus.HistogramVec
}

func newMetricsMiddleware(service string, reqLatencyBuckets []float64, ignore *mux.Router) *metricsMiddleware {
	mw := metricsMiddleware{
		ignore: ignore,
	}

	if reqLatencyBuckets == nil {
		reqLatencyBuckets = prometheus.DefBuckets
	}

	mw.reqLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_request_latency_seconds",
		Help:        "HTTP requests processing latency in seconds.",
		ConstLabels: prometheus.Labels{"service": service},
		Buckets:     reqLatencyBuckets,
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
