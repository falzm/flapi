package middleware

import (
	"net/http"
	"time"

	"github.com/facette/logger"
	"github.com/urfave/negroni"
)

type LoggingMiddlewareConfig struct {
	Logger *logger.Logger
}

type LoggingMiddleware struct {
	*middleware

	log *logger.Logger
}

func NewLoggingMiddleware(config *LoggingMiddlewareConfig, ignore []string) *LoggingMiddleware {
	mw := LoggingMiddleware{
		middleware: newMiddleware(ignore),
		log:        config.Logger,
	}

	return &mw
}

func (mw *LoggingMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	res := rw.(negroni.ResponseWriter)

	if !mw.isIgnored(r) {
		mw.log.Debug("status:%d latency:%s method:%s path:%s", res.Status(), time.Since(start), r.Method, r.URL.Path)
	}
}
