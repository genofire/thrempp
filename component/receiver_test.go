package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

type dummyComp struct {
	Component
	LastPacket stanza.Packet
}

func (d *dummyComp) Connect() (chan stanza.Packet, error) {
	return nil, nil
}
func (d *dummyComp) Send(a stanza.Packet) {
	d.LastPacket = a
}

type dummyXMPP struct {
	xmpp.Sender
	LastPacket stanza.Packet
}

func (d *dummyXMPP) Send(a stanza.Packet) error {
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
	c.handleMessage(s, stanza.IQ{})
	assert.Nil(comp.LastPacket)

	c.handleMessage(s, stanza.Message{})
	_, ok := comp.LastPacket.(stanza.Message)
	assert.True(ok)

	// unsupported iq
	c.handleIQ(s, stanza.IQ{})
	assert.Nil(s.LastPacket)

	c.handleIQ(s, stanza.IQ{
		Attrs: stanza.Attrs{Type: stanza.IQTypeGet},
	})
	assert.NotNil(s.LastPacket)
	iq := s.LastPacket.(stanza.IQ)
	assert.Equal(stanza.IQTypeError, iq.Type)
	assert.Equal("feature-not-implemented", iq.Error.Reason)
	s.LastPacket = nil

	// iq disco info
	c.handleDiscoInfo(s, stanza.IQ{
		Payload: &stanza.DiscoInfo{},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoInfo(s, stanza.IQ{
		Attrs: stanza.Attrs{Type: stanza.IQTypeGet},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoInfo(s, stanza.IQ{
		Attrs:   stanza.Attrs{Type: stanza.IQTypeGet},
		Payload: &stanza.DiscoInfo{},
	})
	assert.NotNil(s.LastPacket)
	iq = s.LastPacket.(stanza.IQ)
	assert.Equal(stanza.IQTypeResult, iq.Type)
	dinfo := iq.Payload.(*stanza.DiscoInfo)
	assert.Equal("monkeyservice", dinfo.Identity[0].Name)
	s.LastPacket = nil

	// iq disco items
	c.handleDiscoItems(s, stanza.IQ{
		Payload: &stanza.DiscoItems{},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoItems(s, stanza.IQ{
		Attrs: stanza.Attrs{Type: stanza.IQTypeGet},
	})
	assert.Nil(s.LastPacket)

	c.handleDiscoItems(s, stanza.IQ{
		Attrs:   stanza.Attrs{Type: stanza.IQTypeGet},
		Payload: &stanza.DiscoItems{},
	})
	assert.NotNil(s.LastPacket)
	iq = s.LastPacket.(stanza.IQ)
	assert.Equal(stanza.IQTypeResult, iq.Type)
	ditems := iq.Payload.(*stanza.DiscoItems)
	assert.Equal("monkeyservice", ditems.Items[0].Name)
	s.LastPacket = nil
}
