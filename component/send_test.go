package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp/stanza"
)

func TestSend(t *testing.T) {
	assert := assert.New(t)

	c := Config{Host: "example.org", XMPPDebug: true}

	// ignoring packet
	p := c.sending(stanza.IQ{})
	assert.Nil(p)

	// send by component host
	p = c.sending(stanza.Message{})
	assert.NotNil(p)
	msg := p.(stanza.Message)
	assert.Equal("example.org", msg.From)

	// send by a user of component
	p = c.sending(stanza.Message{Attrs: stanza.Attrs{From: "threemaid"}})
	assert.NotNil(p)
	msg = p.(stanza.Message)
	assert.Equal("threemaid@example.org", msg.From)
}
