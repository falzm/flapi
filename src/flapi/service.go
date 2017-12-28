package main

import (
	"net/http"

	"github.com/facette/logger"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type service struct {
	server  *http.Server
	metrics *metricsMiddleware
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

	// httpLogs := newlogsMiddleware(log.Context("http"), mwIgnore)
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

	for _, e := range endpoints {
		router.HandleFunc(apiPrefix+e.route, e.handler).
			Methods(e.method)
	}

	router.HandleFunc("/delay", handleDelay).
		Methods("GET", "PUT")

	router.Handle("/metrics", httpMetrics.ServeMetrics()).
		Methods("GET")

	// handlers.Use(httpLogs)
	handlers.Use(httpMetrics)

	handlers.UseHandler(router)

	service.server = &http.Server{
		Addr:    bindAddr,
		Handler: handlers,
	}

	return &service
}

func (s *service) run() error {
	return s.server.ListenAndServe()
}

func (s *service) shutdown() error {
	logger.Notice("shutting down")

	return s.server.Close()
}
