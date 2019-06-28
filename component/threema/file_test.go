package threema

import (
	"testing"

	"gosrc.io/xmpp/stanza"

	"github.com/stretchr/testify/assert"
)

func TestFileToXMPP(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()
	a.threema = &Threema{
		httpUploadURL:  "a",
		httpUploadPath: "/tmp",
	}

	msg, err := a.FileToXMPP("", 1, "jpg", []byte("hallo"))
	assert.NoError(err)
	oob := msg.Extensions[0].(stanza.OOB)
	assert.Equal("a/1.jpg", oob.URL)

	a.threema.httpUploadPath = "/gibt/es/nicht"
	msg, err = a.FileToXMPP("", 1, "jpg", []byte("hallo"))
	assert.Error(err)
	assert.Equal("unable to save file on transport to forward", msg.Body)
}
