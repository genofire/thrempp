package threema

import (
	"errors"
	"testing"

	"github.com/o3ma/o3"
	"github.com/stretchr/testify/assert"
	"gosrc.io/xmpp"
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

func TestReceive(t *testing.T) {
	assert := assert.New(t)

	a := createDummyAccount()

	// receiving/skip error
	p := a.receiving(o3.ReceivedMsg{
		Err: errors.New("dummy"),
	})
	assert.Nil(p)

	// nothing to receiving
	p = a.receiving(o3.ReceivedMsg{})
	assert.Nil(p)
}

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
	p := a.receiving(o3.ReceivedMsg{
		Msg: txtMsg,
	})
	xMSG, ok := p.(xmpp.Message)
	assert.True(ok)
	assert.Equal("Oojoh0Ah", xMSG.Body)

	// receiving/skip text to own id
	session = o3.SessionContext{
		ID: o3.ThreemaID{
			ID:   threemaIDByte,
			Nick: o3.NewPubNick("user"),
		},
	}
	txtMsg, err = o3.NewTextMessage(&session, threemaID, "Aesh8shu")
	assert.NoError(err)
	p = a.receiving(o3.ReceivedMsg{
		Msg: txtMsg,
	})
	assert.Nil(p)
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
	p := a.receiving(o3.ReceivedMsg{
		Msg: drm,
	})
	xMSG, ok := p.(xmpp.Message)
	assert.True(ok)
	rr := xMSG.Extensions[0].(xmpp.ReceiptReceived)
	assert.Equal("im4aeseeh1IbaQui", rr.Id)

	// receiving delivered -> not in cache
	p = a.receiving(o3.ReceivedMsg{
		Msg: drm,
	})
	assert.Nil(p)

	// receiving readed
	drm, err = o3.NewDeliveryReceiptMessage(&session, threemaID, msgID, o3.MSGREAD)
	assert.NoError(err)
	p = a.receiving(o3.ReceivedMsg{
		Msg: drm,
	})
	xMSG, ok = p.(xmpp.Message)
	assert.True(ok)
	cmdd := xMSG.Extensions[0].(xmpp.ChatMarkerDisplayed)
	assert.Equal("im4aeseeh1IbaQui", cmdd.Id)

	// receiving delivered -> not in cache
	p = a.receiving(o3.ReceivedMsg{
		Msg: drm,
	})
	assert.Nil(p)
}
