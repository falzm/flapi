package main

import (
	"fmt"
	"net/http"

	"github.com/facette/httputil"
	"github.com/facette/logger"
	"github.com/falzm/chaos"
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

	httpMetrics, err := newMetricsMiddleware(&metricsMiddlewareConfig{
		service:           "flapi",
		reqLatencyBuckets: config.Metrics.LatencyHistogramBuckets,
	})
	if err != nil {
		return nil, fmt.Errorf("metrics middleware init error: %s", err)
	}

	httpChaos, err := chaos.NewChaos("127.0.0.1:8666")
	if err != nil {
		return nil, fmt.Errorf("chaos middleware init error: %s", err)
	}

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
		router.HandleFunc(e.route, e.handler).
			Methods(e.method)
		log.Debug("registered API endpoint %s %s", e.method, e.route)
	}

	if len(service.endpoints) == 0 {
		log.Warning("no API endpoints registered, check your configuration")
	}

	router.HandleFunc("/", service.handler).
		Methods("GET")

	router.HandleFunc("/metrics", httpMetrics.HandleMetrics).
		Methods("GET")

	// /!\ Middleware chain order matters:
	// - logging/metrics/tracing middleware must be added first, since they measure the whole request process latency
	// - chaos middleware must be added last as it disrupts the request process flow, so instrumentation must
	//   be happen before
	handlers = negroni.New(
		negroni.NewLogger(),
		httpMetrics,
		httpChaos,
	)

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
