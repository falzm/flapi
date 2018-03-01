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
	*middleware

	baseDelay time.Duration
	endpoints map[string]delaySpec
}

var (
	defaultDelayProbability = 1.0
)

func NewDelayMiddleware(config *DelayMiddlewareConfig, ignore []string) *DelayMiddleware {
	mw := DelayMiddleware{
		middleware: newMiddleware(ignore),
		endpoints:  make(map[string]delaySpec),
	}

	return &mw
}

func (mw *DelayMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !mw.isIgnored(r) {
		// Base random delay to avoid flat-lining
		time.Sleep(mw.baseDelay)

		// User-configurable probabilistic delay
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		ds, ok := mw.endpoints[r.Method+r.URL.Path]
		if ok {
			if p := rnd.Float64(); p > 1-ds.probability {
				time.Sleep(ds.duration)
			}
		}
	}

	next(rw, r)
}

func (mw *DelayMiddleware) HandleDelay(rw http.ResponseWriter, r *http.Request) {
	var (
		method string
		route  string
	)

	if mux.Vars(r)["target"] == "endpoint" {
		if method = r.URL.Query().Get("method"); method == "" {
			http.Error(rw, "Missing value for method parameter", http.StatusBadRequest)
			return
		}

		if route = r.URL.Query().Get("route"); route == "" {
			http.Error(rw, "Missing value for route parameter", http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case "GET":
		if mux.Vars(r)["target"] == "base" {
			fmt.Fprintf(rw, "%s\n", mw.baseDelay)
		} else {
			ds, ok := mw.endpoints[method+route]
			if !ok {
				http.Error(rw, "No such endpoint", http.StatusNotFound)
				return
			}

			fmt.Fprintf(rw, "%s (probability: %.1f)\n", ds.duration, ds.probability)
		}

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

		if mux.Vars(r)["target"] == "base" {
			mw.baseDelay = time.Duration(durationValue) * time.Millisecond
		} else {
			probability := defaultDelayProbability
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
