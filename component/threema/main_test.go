package threema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"

	"dev.sum7.eu/genofire/golang-lib/database"
	"dev.sum7.eu/genofire/thrempp/models"
)

func TestThreema(t *testing.T) {
	assert := assert.New(t)

	c, err := NewThreema(map[string]interface{}{})
	assert.NoError(err)
	assert.NotNil(c)

	database.Open(database.Config{
		Type:       "sqlite3",
		Connection: "file::memory:?mode=memory",
	})
	defer database.Close()

	jid := models.JID{
		Local:  "a",
		Domain: "example.org",
	}
	database.Write.Create(&jid)
	database.Write.Create(&models.AccountThreema{
		TID:    []byte("12345678"),
		LSK:    []byte("b"),
		XMPPID: jid.ID,
	})

	//broken
	jid = models.JID{
		Local:  "b",
		Domain: "example.org",
	}
	database.Write.Create(&jid)
	database.Write.Create(&models.AccountThreema{
		TID:    []byte("123"),
		LSK:    []byte("b"),
		XMPPID: jid.ID,
	})

	ch, err := c.Connect()
	assert.NoError(err)
	assert.NotNil(ch)
}
func TestSend(t *testing.T) {
	assert := assert.New(t)

	// test channel
	out := make(chan xmpp.Packet)
	tr := Threema{
		out:        out,
		accountJID: make(map[string]*Account),
	}
	go func() {
		tr.Send(xmpp.Message{
			PacketAttrs: xmpp.PacketAttrs{From: "a@example.org"},
		})
	}()
	p := <-out
	assert.NotNil(p)
	// no account. generate one
	msg := p.(xmpp.Message)
	assert.Contains(msg.Body, "generate")

	// test no answer
	p = tr.send(xmpp.IQ{})
	assert.Nil(p)

	// chat with bot
	p = tr.send(xmpp.Message{
		PacketAttrs: xmpp.PacketAttrs{To: "example.org"},
	})
	assert.NotNil(p)
	msg = p.(xmpp.Message)
	assert.Equal("command not supported", msg.Body)

	// chat with delivier error
	database.Open(database.Config{
		Type:       "sqlite3",
		Connection: "file::memory:?mode=memory",
	})
	defer database.Close()

	jid := models.JID{
		Local:  "a",
		Domain: "example.org",
	}
	database.Write.Create(&jid)
	database.Write.Create(&models.AccountThreema{
		TID:    []byte("12345678"),
		LSK:    []byte("b"),
		XMPPID: jid.ID,
	})

	/* TODO manipulate account to no sendpipe
	_, err := tr.getAccount(&jid)
	assert.NoError(err)

	p = tr.send(xmpp.Message{
		PacketAttrs: xmpp.PacketAttrs{
			From: "a@example.org",
			To:   "12345678@threema.example.org",
		},
	})
	assert.NotNil(p)
	msg = p.(xmpp.Message)
	assert.Equal("command not supported", msg.Body)
	*/
}
