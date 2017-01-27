package main

import "testing"

func TestParseArguments(t *testing.T) {
	_, err := ParseArguments([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Error parsing arguments: %v", err.Error())
	}
}
