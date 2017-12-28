package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/facette/logger"
)

const (
	apiPrefix       = "/api"
	defaultBindAddr = ":8000"
)

var (
	endpoints = map[string]*endpoint{
		"POST" + "/a": newEndpoint("POST", "/a", "OK", http.StatusCreated),
		"GET" + "/a":  newEndpoint("GET", "/a", "A", http.StatusOK),
		"GET" + "/b":  newEndpoint("GET", "/b", "B", http.StatusOK),
		"PUT" + "/c":  newEndpoint("PUT", "/c", "C", http.StatusAccepted),
		"GET" + "/c":  newEndpoint("GET", "/c", "C", http.StatusOK),
	}
	log *logger.Logger

	flagBindAddr string
)

func init() {
	log, _ = logger.NewLogger(logger.FileConfig{Level: "debug"})

	flag.StringVar(&flagBindAddr, "bind-addr", defaultBindAddr, "network [address]:port to bind to")
	flag.Parse()
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
