package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
)

type errorSpec struct {
	statusCode  int
	message     string
	probability float64
}

type ErrorMiddlewareConfig struct {
}

type ErrorMiddleware struct {
	*middleware

	endpoints map[string]errorSpec
}

var (
	defaultErrorProbability = 1.0
	defaultErrorStatusCode  = http.StatusInternalServerError
)

func NewErrorMiddleware(config *ErrorMiddlewareConfig, ignore []string) *ErrorMiddleware {
	mw := ErrorMiddleware{
		middleware: newMiddleware(ignore),
		endpoints:  make(map[string]errorSpec),
	}

	return &mw
}

func (mw *ErrorMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !mw.isIgnored(r) {
		es, ok := mw.endpoints[r.Method+r.URL.Path]
		if ok {
			if p := rand.Float64(); p > 1-es.probability {
				http.Error(rw, es.message, es.statusCode)
				return
			}
		}
	}

	next(rw, r)
}

func (mw *ErrorMiddleware) HandleError(rw http.ResponseWriter, r *http.Request) {
	var (
		method string
		route  string
		err    error
	)

	if method = r.URL.Query().Get("method"); method == "" {
		http.Error(rw, "Missing value for method parameter", http.StatusBadRequest)
		return
	}

	if route = r.URL.Query().Get("route"); route == "" {
		http.Error(rw, "Missing value for route parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		es, ok := mw.endpoints[method+route]
		if !ok {
			http.Error(rw, "No such endpoint", http.StatusNotFound)
			return
		}

		fmt.Fprintf(rw, "%d %q (probability: %.1f)\n", es.statusCode, es.message, es.probability)
		return

	case "PUT":
		statusCode := int64(defaultErrorStatusCode)
		if r.URL.Query().Get("status_code") != "" {
			if statusCode, err = strconv.ParseInt(r.URL.Query().Get("status_code"), 10, 32); err != nil {
				http.Error(rw, "Invalid value for status_code parameter", http.StatusBadRequest)
				return
			}
		}
		if statusCode < 100 || statusCode > 600 {
			http.Error(rw, "Status code parameter value must be 100 < p < 600 ", http.StatusBadRequest)
			return
		}

		message := r.URL.Query().Get("message")

		probability := defaultErrorProbability
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

		mw.endpoints[method+route] = errorSpec{
			statusCode:  int(statusCode),
			message:     message,
			probability: probability,
		}

		rw.WriteHeader(http.StatusNoContent)
		return

	case "DELETE":
		if _, ok := mw.endpoints[method+route]; !ok {
			http.Error(rw, "No such endpoint", http.StatusNotFound)
			return
		}

		delete(mw.endpoints, method+route)

		rw.WriteHeader(http.StatusNoContent)
		return
	}
}
