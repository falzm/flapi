package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type configEndpoint struct {
	Method         string `yaml:"method"`
	Route          string `yaml:"route"`
	ResponseStatus int    `yaml:"response_status"`
	ResponseBody   string `yaml:"response_body"`
}

type config struct {
	Endpoints []*configEndpoint `yaml:"endpoints"`
}

func loadConfig(path string) (*config, error) {
	var c config

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data")
	}

	for _, e := range c.Endpoints {
		if _, err := newEndpoint(e.Method, e.Route, e.ResponseStatus, e.ResponseBody); err != nil {
			return nil, fmt.Errorf("invalid endpoint: %s", err)
		}
	}

	return &c, nil
}

func (c *config) endpoints() map[string]*endpoint {
	endpoints := make(map[string]*endpoint)

	for _, e := range c.Endpoints {
		endpoints[e.Method+e.Route], _ = newEndpoint(e.Method, e.Route, e.ResponseStatus, e.ResponseBody)
	}

	return endpoints
}
