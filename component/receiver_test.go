package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"
)

func TestReceive(t *testing.T) {
	assert := assert.New(t)

	c := Config{Host: "example.org", Type: "monkeyservice", XMPPDebug: true}

	// ignoring packet
	p, _ := c.receiving(xmpp.Handshake{})
	assert.Nil(p)

	// receive presence
	p, _ = c.receiving(xmpp.Presence{})
	assert.Nil(p)

	// message
	p, back := c.receiving(xmpp.Message{})
	assert.False(back)
	assert.NotNil(p)

	// unsupported iq
	p, back = c.receiving(xmpp.IQ{Payload: []xmpp.IQPayload{
		&xmpp.Err{},
	}})
	assert.True(back)
	assert.NotNil(p)
	iq := p.(xmpp.IQ)
	assert.Equal("error", iq.Type)
	assert.Equal("feature-not-implemented", iq.Error.Reason)

	// iq disco info
	p, back = c.receiving(xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
		Payload: []xmpp.IQPayload{
			&xmpp.DiscoInfo{},
		},
	})
	assert.True(back)
	assert.NotNil(p)
	iq = p.(xmpp.IQ)
	assert.Equal("result", iq.Type)
	dinfo := iq.Payload[0].(*xmpp.DiscoInfo)
	assert.Equal("monkeyservice", dinfo.Identity.Name)

	// iq disco items
	p, back = c.receiving(xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
		Payload: []xmpp.IQPayload{
			&xmpp.DiscoItems{},
		},
	})
	assert.True(back)
	assert.NotNil(p)
	iq = p.(xmpp.IQ)
	assert.Equal("result", iq.Type)
	ditems := iq.Payload[0].(*xmpp.DiscoItems)
	assert.Equal("monkeyservice", ditems.Items[0].Name)
}
