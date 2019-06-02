package threema

import (
	"testing"

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
	assert.Equal("a/1.jpg", msg.X.URL)

	a.threema.httpUploadPath = "/gibt/es/nicht"
	msg, err = a.FileToXMPP("", 1, "jpg", []byte("hallo"))
	assert.Error(err)
	assert.Equal("unable to save file on transport to forward", msg.Body)
}
