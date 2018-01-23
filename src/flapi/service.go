package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/facette/logger"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type service struct {
	server    *http.Server
	endpoints map[string]*endpoint
	metrics   *metricsMiddleware
}

func newService(bindAddr string, endpoints map[string]*endpoint) *service {
	var (
		service  service
		handlers *negroni.Negroni
		router   *mux.Router
	)

	handlers = negroni.New()

	mwIgnore := mux.NewRouter()
	mwIgnore.Path("/metrics")
	mwIgnore.Path("/delay")

	httpLogs := newlogsMiddleware(log.Context("http"), mwIgnore)
	httpMetrics := newMetricsMiddleware("flapi",
		map[float64]float64{
			0.5:  0.05,
			0.95: 0.005,
			0.99: 0.001,
		},
		mwIgnore,
	)
	service.metrics = httpMetrics

	router = mux.NewRouter()

	service.endpoints = endpoints
	for _, e := range service.endpoints {
		router.HandleFunc(apiPrefix+e.route, e.handler).
			Methods(e.method)
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

	return &service
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
