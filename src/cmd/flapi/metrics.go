package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type metricsMiddlewareConfig struct {
	service           string
	ignore            *mux.Router
	reqLatencyBuckets []float64
}

type metricsMiddleware struct {
	exporter   *prometheus.Exporter
	ignore     *mux.Router
	reqLatency *stats.MeasureFloat64
	tags       map[string]tag.Key
}

func newMetricsMiddleware(config *metricsMiddlewareConfig) (*metricsMiddleware, error) {
	var (
		err error
		mw  = metricsMiddleware{
			ignore: config.ignore,
			tags:   map[string]tag.Key{},
		}
	)

	if mw.exporter, err = prometheus.NewExporter(prometheus.Options{Namespace: config.service}); err != nil {
		return nil, fmt.Errorf("unable to init Prometheus exporter: %s", err)
	}
	stats.RegisterExporter(mw.exporter)

	if mw.reqLatency, err = stats.NewMeasureFloat64("flapi/measure/http_request_latency",
		"HTTP requests processing latency in seconds",
		"second"); err != nil {
		return nil, fmt.Errorf("unable to create http_request_latency measure: %s", err)
	}

	mw.tags["method"], _ = tag.NewKey("method")
	mw.tags["path"], _ = tag.NewKey("path")
	mw.tags["status"], _ = tag.NewKey("status")

	reqLatencyView, err := stats.NewView(
		"http_request_latency",
		"HTTP requests processing latency in seconds",
		[]tag.Key{mw.tags["method"], mw.tags["path"], mw.tags["status"]},
		mw.reqLatency,
		stats.DistributionAggregation(config.reqLatencyBuckets),
		stats.Cumulative{},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create http_request_latency view: %s", err)
	}

	if err := reqLatencyView.Subscribe(); err != nil {
		return nil, fmt.Errorf("unable to subscribe to http_request_latency view: %s", err)
	}

	stats.SetReportingPeriod(1 * time.Second)

	return &mw, nil
}

func (mw *metricsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	var routeMatch mux.RouteMatch
	if mw.ignore != nil && mw.ignore.Match(r, &routeMatch) {
		return
	}

	res := rw.(negroni.ResponseWriter)

	// TODO: configurable tags
	ctx, err := tag.New(r.Context(),
		tag.Insert(mw.tags["method"], r.Method),
		tag.Insert(mw.tags["path"], r.URL.Path),
		tag.Insert(mw.tags["status"], strconv.Itoa(res.Status())),
	)
	if err != nil {
		log.Error("metricsMiddleware: unable to create tag map: %s", err)
		return
	}

	stats.Record(ctx, mw.reqLatency.M(float64(time.Since(start).Nanoseconds())/1000000000))
}

func (m *metricsMiddleware) ServeMetrics() http.Handler {
	return m.exporter
}
