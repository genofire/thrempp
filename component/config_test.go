package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigStart(t *testing.T) {
	assert := assert.New(t)

	c := Config{}

	// wrong connection
	err := c.Start()
	assert.NotNil(err)

	// correct connection without xmpp server not possible
}
