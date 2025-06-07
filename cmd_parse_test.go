package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizeCMD(t *testing.T) {
	// Basic case - 4 individual tokens
	assert.Equal(t,
		[]string{"download", "the", "whole", "world"},
		tokenizeCMD("download the whole world"),
	)
	// String-enclosed token
	assert.Equal(t,
		[]string{"download", "the ' whole", "world"},
		tokenizeCMD("download \"the ' whole\" world"),
	)
	// Quotation marks don't match (' | ")
	assert.Equal(t, []string(nil), tokenizeCMD("download 'the whole\" world"))
}

func TestParseCMD(t *testing.T) {
	tokens := []string{"download", "world", "--goodbye", "--hello", "world"}
	cmd, flags, args := parseCMD(tokens)
	assert.Equal(t, cmd, "download")
	assert.Equal(t, flags["goodbye"], []string{"true"})
	assert.Equal(t, flags["hello"], []string{"world"})
	assert.Equal(t, args, []string{"world"})
}
