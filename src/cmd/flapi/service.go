package main

import (
	"fmt"
	"net/http"

	"middleware"

	"github.com/facette/httputil"
	"github.com/facette/logger"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type service struct {
	server    *http.Server
	endpoints []*endpoint
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

	router = mux.NewRouter()

	service.endpoints = make([]*endpoint, len(config.Endpoints))
	for i, _ := range config.Endpoints {
		e, err := newEndpoint(
			config.Endpoints[i].Method,
			apiPrefix+config.Endpoints[i].Route,
			config.Endpoints[i].ResponseStatus,
			config.Endpoints[i].ResponseBody,
			config.Endpoints[i].Chain,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint: %s", err)
		}

		service.endpoints[i] = e
		router.HandleFunc(apiPrefix+e.route, e.handler).
			Methods(e.method)
	}

	if len(service.endpoints) == 0 {
		log.Warning("no API endpoints configured, check your configuration")
	}

	router.HandleFunc("/", service.handler).
		Methods("GET")

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

func (s *service) handler(rw http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(rw, s.endpoints, http.StatusOK)
}
