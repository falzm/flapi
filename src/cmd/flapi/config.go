package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	defaultMetricsLatencyHistogramBuckets = []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0}
)

type configMetrics struct {
	LatencyHistogramBuckets []float64 `yaml:"latency_histogram_buckets"`
}

type configEndpointTarget struct {
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
}

type configEndpoint struct {
	Method         string                 `yaml:"method"`
	Route          string                 `yaml:"route"`
	ResponseStatus int                    `yaml:"response_status"`
	ResponseBody   string                 `yaml:"response_body"`
	Chain          []configEndpointTarget `yaml:"chain"`
}

type config struct {
	Metrics   configMetrics     `yaml:metrics`
	Endpoints []*configEndpoint `yaml:"endpoints"`
}

func loadConfig(path string) (*config, error) {
	var c = config{
		Metrics: configMetrics{
			LatencyHistogramBuckets: defaultMetricsLatencyHistogramBuckets,
		},
	}

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

	return &c, nil
}
