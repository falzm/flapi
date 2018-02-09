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
	apiPrefix         = "/api"
	defaultConfigPath = "flapi.yaml"
	defaultBindAddr   = ":8000"
	defaultLogLevel   = "info"
)

var (
	flagConfigPath string
	flagBindAddr   string
	flagLogLevel   string

	log *logger.Logger
)

func init() {
	var err error

	flag.StringVar(&flagConfigPath, "config", defaultConfigPath, "path to configuration file")
	flag.StringVar(&flagBindAddr, "bind-addr", defaultBindAddr, "network [address]:port to bind to")
	flag.StringVar(&flagLogLevel, "log-level", defaultLogLevel, "logging level")
	flag.Parse()

	if log, err = logger.NewLogger(logger.FileConfig{Level: flagLogLevel}); err != nil {
		dieOnError("unable to initialize logger: %s", err)
	}
}

func main() {
	config, err := loadConfig(flagConfigPath)
	if err != nil {
		dieOnError("unable to load configuration: %s", err)
	}

	service, err := newService(flagBindAddr, config)
	if err != nil {
		dieOnError("unable to create service: %s", err)
	}

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

	if err := service.run(); err != nil {
		if err != http.ErrServerClosed {
			log.Error("service: %s", err)
		}
	}

	log.Notice("terminating")
}

func dieOnError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("error: %s\n", format), a...)
	os.Exit(1)
}
