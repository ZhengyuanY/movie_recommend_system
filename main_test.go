package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// A simple example of unit testing a function.
// Adapted from: https://gobyexample.com/testing-and-benchmarking

func TestDummy(t *testing.T) {
	assert.Equal(t, "Hello", "Hello", "pass")
}

