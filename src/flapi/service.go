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
		[]float64{
			0.01, // 10 millisecond
			0.05, // 50 millisecond
			0.1,  // 100 millisecond
			0.5,  // 500 millisecond
			1,    // 1 second
			5,    // 5 seconds
		},
		mwIgnore,
	)
	service.metrics = httpMetrics

	router = mux.NewRouter()

	for _, e := range endpoints {
		router.HandleFunc(apiPrefix+e.route, e.handler).
			Name(apiPrefix + e.route).
			Methods(e.method)
	}

	router.HandleFunc("/delay", handleSet).
		Name("/delay").
		Methods("PUT")

	router.Handle("/metrics", httpMetrics.ServeMetrics()).
		Name("/metrics").
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
