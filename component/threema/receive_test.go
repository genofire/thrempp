package threema

import (
	"testing"

	"github.com/o3ma/o3"
	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp/stanza"
)

const threemaFromID = "87654321"

var threemaFromIDByte o3.IDString

func init() {
	threemaFromIDByte = o3.NewIDString(threemaFromID)
}

func createDummyAccount() Account {
	a := Account{
		deliveredMSG: make(map[uint64]string),
		readedMSG:    make(map[uint64]string),
		xmpp:         make(chan<- stanza.Packet),
	}
	a.TID = make([]byte, len(threemaFromIDByte))
	copy(a.TID, threemaFromIDByte[:])

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
	txtMsg := &o3.TextMessage{
		MessageHeader: &o3.MessageHeader{
			Sender:    threemaFromIDByte,
			Recipient: o3.NewIDString("12345678"),
		},
		Body: "Oojoh0Ah",
	}
	p, err := a.receiving(txtMsg)
	assert.NoError(err)
	assert.NotNil(p)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	assert.Equal("Oojoh0Ah", xMSG.Body)
}

func TestReceiveAudio(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()
	a.threema = &Threema{}

	dataMsg := o3.AudioMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	_, err := a.receiving(dataMsg)
	assert.Error(err)

	a.threema.httpUploadPath = "/tmp"
	dataMsg = o3.AudioMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	_, err = a.receiving(dataMsg)
	assert.Error(err)
}
func TestReceiveImage(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()
	a.threema = &Threema{}

	// receiving image
	dataMsg := o3.ImageMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	_, err := a.receiving(dataMsg)
	assert.Error(err)

	a.threema.httpUploadPath = "/tmp"
	dataMsg = o3.ImageMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	dataMsg = o3.ImageMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	_, err = a.receiving(dataMsg)
	assert.Error(err)
}

func TestReceiveDeliveryReceipt(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// receiving delivered
	msgID := o3.NewMsgID()
	a.deliveredMSG[msgID] = "im4aeseeh1IbaQui"
	a.readedMSG[msgID] = "im4aeseeh1IbaQui"

	drm := &o3.DeliveryReceiptMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
		Status:    o3.MSGDELIVERED,
		MessageID: msgID,
	}
	p, err := a.receiving(drm)
	assert.NoError(err)
	assert.NotNil(p)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	rr := xMSG.Extensions[0].(stanza.ReceiptReceived)
	assert.Equal("im4aeseeh1IbaQui", rr.ID)

	// receiving delivered -> not in cache
	p, err = a.receiving(drm)
	assert.NoError(err)
	assert.Nil(p)

	// receiving readed
	drm = &o3.DeliveryReceiptMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
		MessageID: msgID,
		Status:    o3.MSGREAD,
	}
	p, err = a.receiving(drm)
	assert.NoError(err)
	assert.NotNil(p)
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
	tnm := &o3.TypingNotificationMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
	}
	p, err := a.receiving(tnm)
	assert.NotNil(p)
	assert.NoError(err)
	xMSG, ok := p.(stanza.Message)
	assert.True(ok)
	assert.IsType(stanza.StateInactive{}, xMSG.Extensions[0])

	// receiving composing
	tnm = &o3.TypingNotificationMessage{
		MessageHeader: &o3.MessageHeader{
			Sender: threemaFromIDByte,
		},
		OnOff: 0x1,
	}
	p, err = a.receiving(tnm)
	assert.NotNil(p)
	assert.NoError(err)
	xMSG, ok = p.(stanza.Message)
	assert.True(ok)
	assert.IsType(stanza.StateComposing{}, xMSG.Extensions[0])
}
