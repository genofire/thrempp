package threema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp/stanza"

	"dev.sum7.eu/genofire/golang-lib/database"
	"dev.sum7.eu/sum7/thrempp/models"
)

func TestThreema(t *testing.T) {
	assert := assert.New(t)

	// failed
	c, err := NewThreema(map[string]interface{}{
		"http_upload_path": 3,
	})
	assert.Error(err)
	assert.Nil(c)

	// failed
	c, err = NewThreema(map[string]interface{}{
		"http_upload_url": 3,
	})
	assert.Error(err)
	assert.Nil(c)

	// ---
	c, err = NewThreema(map[string]interface{}{
		"http_upload_url":  "",
		"http_upload_path": "",
	})
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
	out := make(chan stanza.Packet)
	tr := Threema{
		out:        out,
		accountJID: make(map[string]*Account),
		bot:        make(map[string]*Bot),
	}
	go func() {
		tr.Send(stanza.Message{
			Attrs: stanza.Attrs{From: "a@example.org"},
		})
	}()
	p := <-out
	assert.NotNil(p)
	// no account. generate one
	msg := p.(stanza.Message)
	assert.Contains(msg.Body, "generate")

	// test no answer
	p = tr.send(stanza.IQ{})
	assert.Nil(p)

	// chat with bot without sender
	p = tr.send(stanza.Message{
		Attrs: stanza.Attrs{
			To: "example.org",
		},
	})
	assert.Nil(p)

	// chat with bot
	p = tr.send(stanza.Message{
		Attrs: stanza.Attrs{
			From: "a@example.com",
			To:   "example.org",
		},
	})
	assert.NotNil(p)
	msg = p.(stanza.Message)
	assert.Contains(msg.Body, "command not found")

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

	p = tr.send(stanza.Message{
		Attrs: stanza.Attrs{
			From: "a@example.org",
			To:   "12345678@threema.example.org",
		},
	})
	assert.NotNil(p)
	msg = p.(stanza.Message)
	assert.Equal("command not supported", msg.Body)
	*/
}
