package component

import (
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
	AddComponent("a", func(config map[string]interface{}) (Component, error) { return nil, nil })

	Load([]Config{
		{},
		// {Type: "a", Connection: "[::1]:10001"},
	})
}
