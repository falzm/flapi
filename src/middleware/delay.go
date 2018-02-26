package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type delaySpec struct {
	duration    time.Duration
	probability float64
}

type DelayMiddlewareConfig struct {
}

type DelayMiddleware struct {
	ignore    *mux.Router
	baseDelay int
	endpoints map[string]delaySpec
}

var (
	defaultProbability = 1.0
)

func NewDelayMiddleware(config *DelayMiddlewareConfig, ignore *mux.Router) *DelayMiddleware {
	mw := DelayMiddleware{
		ignore:    ignore,
		endpoints: make(map[string]delaySpec),
	}

	return &mw
}

func (mw *DelayMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var routeMatch mux.RouteMatch
	if mw.ignore != nil && mw.ignore.Match(r, &routeMatch) {
		return
	}

	// Base random delay to avoid flat-lining
	time.Sleep(time.Duration(mw.baseDelay) * time.Millisecond)

	// User-configurable probabilistic delay
	ds, ok := mw.endpoints[r.Method+r.URL.Path]
	if ok {
		if p := rand.Float64(); p > 1-ds.probability {
			time.Sleep(ds.duration)
		}
	}

	next(rw, r)
}

func (mw *DelayMiddleware) HandleDelay(rw http.ResponseWriter, r *http.Request) {
	// TODO: make base delay adjustable too (e.g. query string parameter `base`)

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

	switch r.Method {
	case "GET":
		ds, ok := mw.endpoints[method+route]
		if !ok {
			http.Error(rw, "No such endpoint", http.StatusNotFound)
			return
		}

		fmt.Fprintf(rw, "%s (probability: %.1f)\n", ds.duration, ds.probability)
		return

	case "PUT":
		duration := r.URL.Query().Get("duration")
		if duration == "" {
			http.Error(rw, "Missing value for duration parameter", http.StatusBadRequest)
			return
		}

		durationValue, err := strconv.ParseFloat(duration, 64)
		if err != nil {
			http.Error(rw, "Invalid value for duration parameter", http.StatusBadRequest)
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

		mw.endpoints[method+route] = delaySpec{
			duration:    time.Duration(durationValue) * time.Millisecond,
			probability: probability,
		}

		rw.WriteHeader(http.StatusNoContent)
		return
	}
}
