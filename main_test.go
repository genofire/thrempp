package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummy(t *testing.T) {
	assert := assert.New(t)
	assert.Panics(func() {
		main()
	})
}
