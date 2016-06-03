package main

import (
	"net/http"
	"path/filepath"
)

func checkVersion() {
	req, err := http.NewRequest("GET", "https://github.com/hypnoglow/dotbro/releases/latest", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	latestVersion := filepath.Base(resp.Header.Get("Location"))
	currentVersion := "v" + version

	if currentVersion < latestVersion {
		outInfo("Dotbot %s is available, update is recommended.", latestVersion)
	}

}
