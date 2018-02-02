package main

import (
	"fmt"
	"net/http"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/h2non/gentleman.v2"
)

func getDelay(method, route string) {
	res, err := gentleman.New().
		URL(*flapiURL).
		Request().
		Method("GET").
		Path("/delay").
		SetQuery("method", method).
		SetQuery("route", route).
		Send()
	if err != nil {
		kingpin.Fatalf("Request error: %s", err)
	}

	if !res.Ok {
		kingpin.Fatalf("Unexpected server response: %d %s", res.StatusCode, res.String())
	}

	fmt.Println(res.String())
}

func setDelay(method, route, delay, probability string) {
	res, err := gentleman.New().
		URL(*flapiURL).
		Request().
		Method("PUT").
		Path("/delay").
		SetQuery("method", method).
		SetQuery("route", route).
		SetQuery("delay", delay).
		SetQuery("probability", probability).
		Send()
	if err != nil {
		kingpin.Fatalf("Request error: %s", err)
	}

	if res.StatusCode != http.StatusNoContent {
		kingpin.Fatalf("Unexpected server response: %d %s", res.StatusCode, res.String())
	}

	fmt.Println("OK")
}
