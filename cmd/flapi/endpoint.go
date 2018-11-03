package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/facette/httputil"
)

type endpointTarget struct {
	client *http.Client
	method string
	url    *url.URL
}

func (e *endpointTarget) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"method": e.method,
		"url":    e.url.String(),
	})
}

func (e *endpointTarget) request(ctx context.Context) (*http.Response, error) {
	e.client = http.DefaultClient

	log.Debug("requesting target endpoint: %s %s", e.method, e.url.String())

	req, err := http.NewRequest(e.method, e.url.String(), nil)
	if err != nil {
		return nil, err
	}

	return e.client.Do(req.WithContext(ctx))
}

type endpoint struct {
	method          string
	route           string
	responseStatus  int
	responseHeaders map[string]string
	responseBody    string
	targets         []endpointTarget
}

func newEndpoint(method, route string, responseStatus int, responseBody string,
	targets []configEndpointTarget) (*endpoint, error) {
	var (
		e   endpoint
		err error
	)

	if method == "" {
		return nil, fmt.Errorf("method not specified")
	}
	e.method = method

	if route == "" {
		return nil, fmt.Errorf("route not specified")
	}
	e.route = route

	if (responseStatus < 100 || responseStatus > 599) && targets == nil {
		return nil, fmt.Errorf("invalid response status code")
	}
	e.responseStatus = responseStatus

	e.responseHeaders = map[string]string{
		"Version": version,
		"Host":    hostname,
	}
	for _, env := range os.Environ() {
		sub := strings.Split(env, "=")
		if strings.HasPrefix(sub[0], "FLAPI_") {
			e.responseHeaders[strings.TrimPrefix(sub[0], "FLAPI_")] = sub[1]
		}
	}

	e.responseBody = responseBody

	if targets != nil {
		e.targets = make([]endpointTarget, len(targets))
		for i := range targets {
			if targets[i].Method == "" {
				return nil, fmt.Errorf("invalid endpoint chain: missing remote endpoint method")
			}
			e.targets[i].method = targets[i].Method

			if targets[i].URL == "" {
				return nil, fmt.Errorf("invalid endpoint chain: missing remote endpoint URL")
			}

			if e.targets[i].url, err = url.Parse(targets[i].URL); err != nil {
				return nil, fmt.Errorf("invalid endpoint chain: URL: %s", err)
			}
		}
	}

	return &e, nil
}

func (e *endpoint) handler(rw http.ResponseWriter, r *http.Request) {
	for k, v := range e.responseHeaders {
		rw.Header().Set("X-Flapi-"+k, v)
	}

	if e.targets == nil {
		rw.WriteHeader(e.responseStatus)
		fmt.Fprintf(rw, "%s\n", e.responseBody)
	} else {
		finalStatus := http.StatusOK
		targetResponses := make([]string, len(e.targets))

		// TODO: request targets concurrently with goroutines
		for i, t := range e.targets {
			res, err := t.request(r.Context())
			if err != nil {
				finalStatus = http.StatusInternalServerError
				targetResponses[i] = fmt.Sprintf("error: %s", err)
				continue
			}
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				finalStatus = http.StatusInternalServerError
				continue
			}

			targetResponses[i] = fmt.Sprintf("HTTP %s: %s", res.Status, strings.TrimSpace(string(data)))
		}

		httputil.WriteJSON(rw, targetResponses, finalStatus)
	}
}

func (e *endpoint) MarshalJSON() ([]byte, error) {
	je := map[string]interface{}{
		"method": e.method,
		"route":  e.route,
	}

	if e.targets != nil {
		je["targets"] = e.targets
	} else {
		je["response_status"] = e.responseStatus
		je["response_body"] = e.responseBody
	}

	return json.Marshal(je)
}
