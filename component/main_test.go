package component

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddComponent(t *testing.T) {
	assert := assert.New(t)

	AddComponent("a", func(config map[string]interface{}) (Component, error) { return nil, nil })

	assert.NotNil(components["a"])
	assert.Len(components, 1)
}

func TestLoad(t *testing.T) {
	assert := assert.New(t)

	AddComponent("error", func(config map[string]interface{}) (Component, error) {
		return nil, errors.New("dummy")
	})
	// load correct
	Load([]Config{
		{},
	})

	// error on component
	assert.Panics(func() {
		Load([]Config{
			{Type: "error", Connection: "[::1]:10001"},
		})
	})

	// error on connect
	assert.Panics(func() {
		Load([]Config{
			{Type: "a", Connection: "[::1]:10001"},
		})
	})
}
