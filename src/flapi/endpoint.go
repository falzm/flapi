package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type endpoint struct {
	method         string
	route          string
	responseStatus int
	responseBody   string
	delay          time.Duration
	probability    float64
}

var defaultProbability = 0.0

func newEndpoint(method, route string, responseStatus int, responseBody string) (*endpoint, error) {
	if method == "" {
		return nil, fmt.Errorf("method not specified")
	}

	if route == "" {
		return nil, fmt.Errorf("route not specified")
	}

	if responseStatus < 100 || responseStatus > 599 {
		return nil, fmt.Errorf("invalid response status code")
	}

	return &endpoint{
		method:         method,
		route:          route,
		responseStatus: responseStatus,
		responseBody:   responseBody,
		probability:    defaultProbability,
	}, nil
}

func (e *endpoint) handler(rw http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)

	if p := rand.Float64(); p > e.probability {
		time.Sleep(e.delay)
	}

	rw.WriteHeader(e.responseStatus)

	fmt.Fprintf(rw, "%s\n", e.responseBody)
}

func (e *endpoint) setDelay(delay time.Duration, probability float64) {
	e.delay = delay
	e.probability = probability
}
