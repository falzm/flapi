package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/falzm/chaos"

	"github.com/facette/logger"
)

const (
	apiPrefix         = "/api"
	defaultConfigPath = "flapi.yaml"
	defaultBindAddr   = ":8000"
	defaultLogLevel   = "info"
)

var (
	version   string
	buildDate string

	flagBindAddr      string
	flagChaosBindAddr string
	flagConfigPath    string
	flagHelp          bool
	flagLogLevel      string
	flagVersion       bool

	log *logger.Logger
)

func init() {
	var err error

	flag.BoolVar(&flagHelp, "help", false, "display this help and exit")
	flag.BoolVar(&flagVersion, "version", false, "display version and exit")
	flag.StringVar(&flagBindAddr, "bind-addr", defaultBindAddr, "HTTP server network [address]:port to bind to")
	flag.StringVar(&flagChaosBindAddr, "chaos-bind-addr", chaos.DefaultBindAddr, "chaos management HTTP server network [address]:port to bind to")
	flag.StringVar(&flagConfigPath, "config", defaultConfigPath, "path to configuration file")
	flag.StringVar(&flagLogLevel, "log-level", defaultLogLevel, "logging level")
	flag.Parse()

	if log, err = logger.NewLogger(logger.FileConfig{Level: flagLogLevel}); err != nil {
		dieOnError("unable to initialize logger: %s", err)
	}
}

func main() {
	if flagHelp {
		printUsage(os.Stdout)
		os.Exit(0)
	} else if flagVersion {
		printVersion(version, buildDate)
		os.Exit(0)
	}

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
	log.Debug("chaos management listening on %s", flagChaosBindAddr)

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

func printUsage(output io.Writer) {
	fmt.Fprintf(output, "Usage: %s [options]", path.Base(os.Args[0]))
	fmt.Fprint(output, "\n\nOptions:\n")

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(output, "   -%s  %s (default: %q)\n", f.Name, f.Usage, f.DefValue)
	})

	os.Exit(2)
}

func printVersion(version, buildDate string) {
	fmt.Printf("%s version %s, built on %s\nGo version: %s (%s)\n",
		path.Base(os.Args[0]),
		version,
		buildDate,
		runtime.Version(),
		runtime.Compiler,
	)
}
