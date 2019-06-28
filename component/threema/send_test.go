package threema

import (
	"testing"

	"github.com/o3ma/o3"
	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp/stanza"
)

func TestAccountSend(t *testing.T) {
	assert := assert.New(t)

	send := make(chan o3.Message)
	a := Account{
		Session:      o3.NewSessionContext(o3.ThreemaID{ID: o3.NewIDString("43218765")}),
		send:         send,
		deliveredMSG: make(map[uint64]string),
		readedMSG:    make(map[uint64]string),
	}

	go func() {
		a.Send("a", stanza.Message{
			Attrs: stanza.Attrs{From: "a@example.org"},
			Body:  "ohz8kai0ohNgohth",
		})
	}()
	p := <-send
	assert.NotNil(p)
	msg := p.(o3.TextMessage)
	assert.Contains(msg.Text(), "ohz8kai0ohNgohth")

	// test error
	err := a.Send("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.ReceiptReceived{ID: "blub"},
		},
	})
	assert.Error(err)
}

func TestAccountSendingDeliviery(t *testing.T) {
	assert := assert.New(t)

	a := Account{
		Session: o3.NewSessionContext(o3.ThreemaID{ID: o3.NewIDString("43218765")}),
	}

	// test error - threema send only int ids
	msg, err := a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.ReceiptReceived{ID: "blub"},
		},
	})
	assert.Error(err)
	assert.Nil(msg)

	// test delivered
	msg, err = a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.MarkReceived{ID: "3"},
		},
	})
	assert.NoError(err)
	drm, ok := msg.(o3.DeliveryReceiptMessage)
	assert.True(ok)
	assert.Equal(o3.MSGDELIVERED, drm.Status())

	// test read
	msg, err = a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.MarkDisplayed{ID: "5"},
		},
	})
	assert.NoError(err)
	drm, ok = msg.(o3.DeliveryReceiptMessage)
	assert.True(ok)
	assert.Equal(o3.MSGREAD, drm.Status())
}
func TestSendTyping(t *testing.T) {
	assert := assert.New(t)

	a := Account{
		Session:      o3.NewSessionContext(o3.ThreemaID{ID: o3.NewIDString("43218765")}),
		deliveredMSG: make(map[uint64]string),
		readedMSG:    make(map[uint64]string),
	}

	// skip typing messae
	msg, err := a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.StateComposing{},
		},
	})
	assert.NoError(err)
	assert.Nil(msg)

	// skip gone
	msg, err = a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.StateActive{},
			&stanza.StateGone{},
			&stanza.StateInactive{},
			&stanza.StatePaused{},
		},
	})
	assert.NoError(err)
	assert.Nil(msg)

	// skip gone
	msg, err = a.sending("a", stanza.Message{
		Attrs: stanza.Attrs{From: "a@example.org"},
		Extensions: []stanza.MsgExtension{
			&stanza.StateActive{},
			&stanza.StateComposing{},
			&stanza.StateGone{},
			&stanza.StateInactive{},
			&stanza.StatePaused{},
		},
		Body: "hi",
	})
	assert.NoError(err)
	assert.NotNil(msg)
	o3msg := msg.(o3.TextMessage)
	assert.Contains(o3msg.Text(), "hi")
}
