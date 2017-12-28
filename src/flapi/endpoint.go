package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type endpoint struct {
	method   string
	route    string
	response string
	code     int
	delay    time.Duration
}

func newEndpoint(method, route, response string, code int) *endpoint {
	return &endpoint{
		method:   method,
		route:    route,
		response: response,
		code:     code,
	}
}

func (e *endpoint) handler(rw http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(1+rand.Intn(10))*time.Millisecond + e.delay)

	rw.WriteHeader(e.code)

	fmt.Fprintf(rw, "%s\n", e.response)
}

func handleSet(rw http.ResponseWriter, r *http.Request) {
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

	e, ok := endpoints[method+route]
	if !ok {
		http.Error(rw, "No such endpoint", http.StatusNotFound)
		return
	}

	e.setDelay(time.Duration(delayValue) * time.Millisecond)

	log.Info("delay for endpoint %s:%s adjusted to %s\n", e.method, e.route, e.delay)

	rw.WriteHeader(http.StatusNoContent)
}

func (e *endpoint) setDelay(delay time.Duration) {
	e.delay = delay
}
