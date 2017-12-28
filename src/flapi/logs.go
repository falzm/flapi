package main

import (
	"net/http"
	"time"

	"github.com/facette/logger"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type logsMiddleware struct {
	ignore *mux.Router
	log    *logger.Logger
}

func newlogsMiddleware(logger *logger.Logger, ignore *mux.Router) *logsMiddleware {
	mw := logsMiddleware{
		log:    logger,
		ignore: ignore,
	}

	return &mw
}

func (mw *logsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	var routeMatch mux.RouteMatch
	if mw.ignore != nil && mw.ignore.Match(r, &routeMatch) {
		return
	}

	res := rw.(negroni.ResponseWriter)

	mw.log.Info("status:%d latency:%s method:%s path:%s", res.Status(), time.Since(start), r.Method, r.URL.Path)
}
