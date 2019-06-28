package threema

import (
	"testing"

	"github.com/o3ma/o3"
	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp/stanza"
)

const threemaID = "87654321"

var threemaIDByte o3.IDString

func init() {
	threemaIDByte = o3.NewIDString(threemaID)
}

func createDummyAccount() Account {
	a := Account{
		deliveredMSG: make(map[uint64]string),
		readedMSG:    make(map[uint64]string),
	}
	a.TID = make([]byte, len(threemaIDByte))
	copy(a.TID, threemaIDByte[:])

	return a
}

/*
func TestReceive(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// nothing to receiving
	p, err := a.receiving(nil)
	assert.Nil(p)
	assert.Error(err)
}
*/
func TestReceiveText(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// receiving text
	session := o3.SessionContext{
		ID: o3.ThreemaID{
			ID:   o3.NewIDString("12345678"),
			Nick: o3.NewPubNick("user"),
		},
	}
	txtMsg, err := o3.NewTextMessage(&session, threemaID, "Oojoh0Ah")
	assert.NoError(err)
	p, err := a.receiving(txtMsg)
	assert.NoError(err)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	assert.Equal("Oojoh0Ah", xMSG.Body)
}

func TestReceiveAudio(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()
	a.threema = &Threema{}

	/* receiving image
	session := o3.SessionContext{
		ID: o3.ThreemaID{
			ID:   o3.NewIDString("12345678"),
			Nick: o3.NewPubNick("user"),
		},
	}*/
	dataMsg := o3.AudioMessage{}
	_, err := a.receiving(dataMsg)
	assert.Error(err)

	a.threema.httpUploadPath = "/tmp"
	dataMsg = o3.AudioMessage{}
	_, err = a.receiving(dataMsg)
	assert.Error(err)
}
func TestReceiveImage(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()
	a.threema = &Threema{}

	/* receiving image
	session := o3.SessionContext{
		ID: o3.ThreemaID{
			ID:   o3.NewIDString("12345678"),
			Nick: o3.NewPubNick("user"),
		},
	}*/
	imgMsg := o3.ImageMessage{}
	_, err := a.receiving(imgMsg)
	assert.Error(err)

	a.threema.httpUploadPath = "/tmp"
	imgMsg = o3.ImageMessage{}
	_, err = a.receiving(imgMsg)
	assert.Error(err)
}

func TestReceiveDeliveryReceipt(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// receiving delivered
	session := o3.SessionContext{
		ID: o3.ThreemaID{
			ID:   o3.NewIDString("12345678"),
			Nick: o3.NewPubNick("user"),
		},
	}
	msgID := o3.NewMsgID()
	a.deliveredMSG[msgID] = "im4aeseeh1IbaQui"
	a.readedMSG[msgID] = "im4aeseeh1IbaQui"

	drm, err := o3.NewDeliveryReceiptMessage(&session, threemaID, msgID, o3.MSGDELIVERED)
	assert.NoError(err)
	p, err := a.receiving(drm)
	assert.NoError(err)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	rr := xMSG.Extensions[0].(stanza.ReceiptReceived)
	assert.Equal("im4aeseeh1IbaQui", rr.ID)

	// receiving delivered -> not in cache
	p, err = a.receiving(drm)
	assert.NoError(err)
	assert.Nil(p)

	// receiving readed
	drm, err = o3.NewDeliveryReceiptMessage(&session, threemaID, msgID, o3.MSGREAD)
	assert.NoError(err)
	p, err = a.receiving(drm)
	assert.NoError(err)
	xMSG, ok = p.(stanza.Message)
	assert.True(ok)
	cmdd := xMSG.Extensions[0].(stanza.MarkDisplayed)
	assert.Equal("im4aeseeh1IbaQui", cmdd.ID)

	// receiving delivered -> not in cache
	p, err = a.receiving(drm)
	assert.NoError(err)
	assert.Nil(p)
}
func TestReceiveTyping(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// receiving inactive
	tnm := o3.TypingNotificationMessage{}
	p, err := a.receiving(tnm)
	assert.NoError(err)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	assert.IsType(stanza.StateInactive{}, xMSG.Extensions[0])

	// receiving composing
	tnm = o3.TypingNotificationMessage{}
	tnm.OnOff = 0x1
	p, err = a.receiving(tnm)
	assert.NoError(err)
	xMSG, ok = p.(stanza.Message)
	assert.True(ok)
	assert.IsType(stanza.StateComposing{}, xMSG.Extensions[0])
}
