package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/facette/logger"
)

const (
	apiPrefix       = "/api"
	defaultBindAddr = ":8000"
	defaultLogLevel = "info"
)

var (
	endpoints = map[string]*endpoint{
		"POST" + "/a": newEndpoint("POST", "/a", "OK", http.StatusCreated),
		"GET" + "/a":  newEndpoint("GET", "/a", "A", http.StatusOK),
		"GET" + "/b":  newEndpoint("GET", "/b", "B", http.StatusOK),
		"PUT" + "/c":  newEndpoint("PUT", "/c", "C", http.StatusAccepted),
		"GET" + "/c":  newEndpoint("GET", "/c", "C", http.StatusOK),
	}

	flagBindAddr string
	flagLogLevel string

	log *logger.Logger
)

func init() {
	var err error

	flag.StringVar(&flagBindAddr, "bind-addr", defaultBindAddr, "network [address]:port to bind to")
	flag.StringVar(&flagLogLevel, "log-level", defaultLogLevel, "logging level")
	flag.Parse()

	if log, err = logger.NewLogger(logger.FileConfig{Level: flagLogLevel}); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to initialize logger: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	service := newService(flagBindAddr, endpoints)

	// Handle service signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				service.shutdown()
			}
		}
	}()

	log.Notice("starting")

	log.Debug("listening on %s", flagBindAddr)

	service.run()

	log.Notice("terminating")
}
