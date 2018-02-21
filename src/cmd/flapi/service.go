package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	mwIgnore := mux.NewRouter()
	mwIgnore.Path("/metrics")
	mwIgnore.Path("/delay")

	httpLogs := middleware.NewLoggingMiddleware(log.Context("http"), mwIgnore)
	httpMetrics, err := middleware.NewMetricsMiddleware(&middleware.MetricsMiddlewareConfig{
		Service:           "flapi",
		Ignore:            mwIgnore,
		ReqLatencyBuckets: config.Metrics.LatencyHistogramBuckets,
	})
	if err != nil {
		return nil, fmt.Errorf("metrics middleware init error: %s", err)
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

	router.HandleFunc("/delay", service.handleDelay).
		Methods("GET", "PUT")

	router.Handle("/metrics", httpMetrics.ServeMetrics()).
		Methods("GET")

	handlers.Use(httpLogs)
	handlers.Use(httpMetrics)

	handlers.UseHandler(router)

	service.server = &http.Server{
		Addr:    bindAddr,
		Handler: handlers,
	}

	return &service, nil
}

func (s *service) handleDelay(rw http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")
	if method == "" {
		http.Error(rw, "Missing value for method parameter", http.StatusBadRequest)
		return
	}

	route := r.URL.Query().Get("route")
	if route == "" {
		http.Error(rw, "Missing value for route parameter", http.StatusBadRequest)
		return
	}

	e, ok := s.endpoints[method+route]
	if !ok {
		http.Error(rw, "No such endpoint", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		fmt.Fprintf(rw, "%s (probability: %.1f)\n", e.delay, e.probability)
		return

	case "PUT":
		delay := r.URL.Query().Get("delay")
		if delay == "" {
			http.Error(rw, "Missing value for delay parameter", http.StatusBadRequest)
			return
		}

		delayValue, err := strconv.ParseFloat(delay, 64)
		if err != nil {
			http.Error(rw, "Invalid value for delay parameter", http.StatusBadRequest)
			return
		}

		probability := defaultProbability
		if r.URL.Query().Get("probability") != "" {
			probability, err = strconv.ParseFloat(r.URL.Query().Get("probability"), 64)
			if err != nil {
				http.Error(rw, "Invalid value for probability parameter", http.StatusBadRequest)
				return
			}
		}
		if probability < 0 || probability > 1 {
			http.Error(rw, "Probability parameter value must be 0 < p < 1 ", http.StatusBadRequest)
			return
		}

		e.setDelay(time.Duration(delayValue)*time.Millisecond, probability)
		log.Info("delay for endpoint %s:%s adjusted to %s with probability %.1f",
			e.method,
			e.route,
			e.delay,
			e.probability,
		)

		rw.WriteHeader(http.StatusNoContent)
		return
	}
}

func (s *service) run() error {
	return s.server.ListenAndServe()
}

func (s *service) shutdown() error {
	logger.Notice("shutting down")

	return s.server.Close()
}
