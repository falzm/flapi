package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultURL = "http://localhost:8000"
)

var (
	flapiURL = kingpin.Flag("url", "flapi server URL").Default(defaultURL).String()

	delayCmd = kingpin.Command("delay", "Manage endpoint delay.")

	delayGetCmd          = delayCmd.Command("get", "Get endpoint current delay value.")
	delayGetCmdArgMethod = delayGetCmd.Arg("method", "endpoint method").Required().String()
	delayGetCmdArgRoute  = delayGetCmd.Arg("route", "endpoint route").Required().String()

	delaySetCmd                = delayCmd.Command("set", "Set endpoint delay.")
	delaySetCmdFlagProbability = delaySetCmd.
					Flag("probability", "delay probability (0.0 >= p <= 1.0)").
					Default("1.0").
					Short('p').
					String()
	delaySetCmdArgMethod = delaySetCmd.Arg("method", "endpoint method").Required().String()
	delaySetCmdArgRoute  = delaySetCmd.Arg("route", "endpoint route").Required().String()
	delaySetCmdArgDelay  = delaySetCmd.Arg("delay", "delay in milliseconds").Required().String()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')

	switch kingpin.Parse() {
	case delayGetCmd.FullCommand():
		getDelay(*delayGetCmdArgMethod, *delayGetCmdArgRoute)

	case delaySetCmd.FullCommand():
		setDelay(*delaySetCmdArgMethod, *delaySetCmdArgRoute, *delaySetCmdArgDelay, *delaySetCmdFlagProbability)
	}
}
