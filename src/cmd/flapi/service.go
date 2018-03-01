package main

import (
	"fmt"
	"net/http"

	"middleware"

	"github.com/facette/logger"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type service struct {
	server    *http.Server
	endpoints map[string]*endpoint
}

func newService(bindAddr string, config *config) (*service, error) {
	var (
		service  service
		handlers *negroni.Negroni
		router   *mux.Router
		err      error
	)

	handlers = negroni.New()

	httpLogging := middleware.NewLoggingMiddleware(&middleware.LoggingMiddlewareConfig{Logger: log.Context("http")},
		[]string{"/metrics", "/delay", "/error"})

	httpMetrics, err := middleware.NewMetricsMiddleware(&middleware.MetricsMiddlewareConfig{
		Service:           "flapi",
		ReqLatencyBuckets: config.Metrics.LatencyHistogramBuckets,
	},
		[]string{"/metrics", "/delay", "/error"})
	if err != nil {
		return nil, fmt.Errorf("metrics middleware init error: %s", err)
	}

	httpDelay := middleware.NewDelayMiddleware(&middleware.DelayMiddlewareConfig{},
		[]string{"/metrics", "/delay", "/error"})

	httpError := middleware.NewErrorMiddleware(&middleware.ErrorMiddlewareConfig{},
		[]string{"/metrics", "/delay", "/error"})

	if config.Trace.JaegerEndpoint != "" {
		httpTrace, err := middleware.NewTraceMiddleware(&middleware.TraceMiddlewareConfig{
			Service:        "flapi",
			JaegerEndpoint: config.Trace.JaegerEndpoint,
		},
			[]string{"/metrics", "/delay", "/error"})
		if err != nil {
			return nil, fmt.Errorf("trace middleware init error: %s", err)
		}
		handlers.Use(httpTrace)
	}

	router = mux.NewRouter()

	service.endpoints = make(map[string]*endpoint)
	for _, e := range config.Endpoints {
		if service.endpoints[e.Method+e.Route], err = newEndpoint(
			e.Method,
			e.Route,
			e.ResponseStatus,
			e.ResponseBody,
			e.Chain,
		); err != nil {
			return nil, fmt.Errorf("invalid endpoint: %s", err)
		}
		router.HandleFunc(apiPrefix+e.Route, service.endpoints[e.Method+e.Route].handler).
			Methods(e.Method)
	}

	router.HandleFunc("/delay/{target}", httpDelay.HandleDelay).
		Methods("GET", "PUT", "DELETE")

	router.HandleFunc("/error", httpError.HandleError).
		Methods("GET", "PUT", "DELETE")

	router.HandleFunc("/metrics", httpMetrics.HandleMetrics).
		Methods("GET")

	// /!\ Middleware chain order matters:
	// - logging/metrics middleware have to be added first, since they measure the whole request process latency
	// - error middleware has to be added last as it interrupts the request process latency, and instrumentation must
	//   be performed before
	handlers.Use(httpLogging)
	handlers.Use(httpMetrics)
	handlers.Use(httpDelay)
	handlers.Use(httpError)

	handlers.UseHandler(router)

	service.server = &http.Server{
		Addr:    bindAddr,
		Handler: handlers,
	}

	return &service, nil
}

func (s *service) run() error {
	return s.server.ListenAndServe()
}

func (s *service) shutdown() error {
	logger.Notice("shutting down")

	return s.server.Close()
}
