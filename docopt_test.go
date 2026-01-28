package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArguments(t *testing.T) {
	_, err := ParseArguments([]string{"--quiet"})
	require.NoError(t, err)
}
