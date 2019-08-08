package threema

import (
	"testing"

	"dev.sum7.eu/genofire/golang-lib/database"
	"github.com/stretchr/testify/assert"

	"dev.sum7.eu/sum7/thrempp/models"
)

func TestBot(t *testing.T) {
	assert := assert.New(t)

	b := Bot{
		jid: &models.JID{},
		threema: &Threema{
			bot: make(map[string]*Bot),
		},
	}

	msg := b.Handle("help")
	assert.NotEqual("", msg)

	// getAccount
	a, err := b.getAccount()
	assert.Error(err)
	assert.Nil(a)
}

func TestGetBot(t *testing.T) {
	assert := assert.New(t)

	tr := Threema{
		bot: make(map[string]*Bot),
	}
	jid := &models.JID{
		Local:  "a",
		Domain: "example.org",
	}
	//
	b := tr.getBot(jid)
	assert.NotNil(b)

	// getBot from cache
	b = tr.getBot(jid)
	assert.NotNil(b)

	// reset cache + test jid db
	tr.bot = make(map[string]*Bot)
	database.Open(database.Config{
		Type:       "sqlite3",
		Connection: "file::memory:?mode=memory",
	})
	defer database.Close()
	b = tr.getBot(jid)
	assert.NotNil(b)
}
func TestBotGenerate(t *testing.T) {
	assert := assert.New(t)

	threema := &Threema{
		bot:        make(map[string]*Bot),
		accountJID: make(map[string]*Account),
	}

	b := threema.getBot(&models.JID{
		Local:  "generate",
		Domain: "example.org",
	})

	// failed to generate without db
	msg := b.Handle("generate")
	assert.Equal("failed to create a threema account", msg)

	database.Open(database.Config{
		Type:       "sqlite3",
		Connection: "file::memory:?mode=memory",
	})
	threema = &Threema{
		bot:        make(map[string]*Bot),
		accountJID: make(map[string]*Account),
	}
	b = threema.getBot(&models.JID{
		Local:  "generate",
		Domain: "example.org",
	})

	// generate
	msg = b.Handle("generate")
	assert.Contains(msg, "threema account with id")

	// alread generated
	msg = b.Handle("generate")
	assert.Contains(msg, "threema account with id")

	database.Close()
}
