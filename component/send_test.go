package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"
)

func TestSend(t *testing.T) {
	assert := assert.New(t)

	c := Config{Host: "example.org", XMPPDebug: true}

	// ignoring packet
	p := c.sending(xmpp.IQ{})
	assert.Nil(p)

	// send by component host
	p = c.sending(xmpp.Message{})
	assert.NotNil(p)
	msg := p.(xmpp.Message)
	assert.Equal("example.org", msg.PacketAttrs.From)

	// send by a user of component
	p = c.sending(xmpp.Message{PacketAttrs: xmpp.PacketAttrs{From: "threemaid"}})
	assert.NotNil(p)
	msg = p.(xmpp.Message)
	assert.Equal("threemaid@example.org", msg.PacketAttrs.From)
}
