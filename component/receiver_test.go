package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"
)

type dummyComp struct {
	Component
	LastPacket xmpp.Packet
}

func (d *dummyComp) Connect() (chan xmpp.Packet, error) {
	return nil, nil
}
func (d *dummyComp) Send(a xmpp.Packet) {
	d.LastPacket = a
}

type dummyXMPP struct {
	xmpp.Sender
	LastPacket xmpp.Packet
}

func (d *dummyXMPP) Send(a xmpp.Packet) error {
	d.LastPacket = a
	return nil
}

func TestReceive(t *testing.T) {
	assert := assert.New(t)
	s := &dummyXMPP{}

	comp := &dummyComp{}
	c := Config{
		Host:      "example.org",
		Type:      "monkeyservice",
		XMPPDebug: true,
		comp:      comp,
	}

	// message
	c.handleMessage(s, xmpp.IQ{})
	assert.Nil(comp.LastPacket)

	c.handleMessage(s, xmpp.Message{})
	_, ok := comp.LastPacket.(xmpp.Message)
	assert.True(ok)

	// unsupported iq
	c.handleIQ(s, xmpp.IQ{})
	assert.Nil(s.LastPacket)

	c.handleIQ(s, xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
	})
	assert.NotNil(s.LastPacket)
	iq := s.LastPacket.(xmpp.IQ)
	assert.Equal("error", iq.Type)
	assert.Equal("feature-not-implemented", iq.Error.Reason)
	s.LastPacket = nil

	// iq disco info
	c.handleDiscoInfo(s, xmpp.IQ{
		Payload: &xmpp.DiscoInfo{},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoInfo(s, xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoInfo(s, xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
		Payload:     &xmpp.DiscoInfo{},
	})
	assert.NotNil(s.LastPacket)
	iq = s.LastPacket.(xmpp.IQ)
	assert.Equal("result", iq.Type)
	dinfo := iq.Payload.(*xmpp.DiscoInfo)
	assert.Equal("monkeyservice", dinfo.Identity.Name)
	s.LastPacket = nil

	// iq disco items
	c.handleDiscoItems(s, xmpp.IQ{
		Payload: &xmpp.DiscoItems{},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoItems(s, xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoItems(s, xmpp.IQ{
		PacketAttrs: xmpp.PacketAttrs{Type: "get"},
		Payload:     &xmpp.DiscoItems{},
	})
	assert.NotNil(s.LastPacket)
	iq = s.LastPacket.(xmpp.IQ)
	assert.Equal("result", iq.Type)
	ditems := iq.Payload.(*xmpp.DiscoItems)
	assert.Equal("monkeyservice", ditems.Items[0].Name)
	s.LastPacket = nil
}
