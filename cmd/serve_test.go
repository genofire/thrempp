package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	assert := assert.New(t)

	// fail on open file
	RootCmd.SetArgs([]string{"serve", "--config", "a"})
	assert.Panics(func() {
		Execute()
	})

	// run
	RootCmd.SetArgs([]string{"serve", "--config", "../config_example.toml"})

	assert.Panics(func() {
		Execute()
	})
}
