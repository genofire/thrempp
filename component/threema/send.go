package threema

import (
	"strconv"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp"
)

func (a *Account) Send(to string, msg xmpp.Message) error {
	m, err := a.sending(to, msg)
	if err != nil {
		return err
	}
	if m != nil {
		a.send <- m
	}
	return nil
}
func (a *Account) sending(to string, msg xmpp.Message) (o3.Message, error) {
	// handle delivered / readed
	msgID := ""
	readed := false
	composing := false
	state := false
	for _, el := range msg.Extensions {
		switch ex := el.(type) {
		case xmpp.StateComposing:
			composing = true
			state = true
		case xmpp.StateInactive:
			state = true
		case xmpp.StateActive:
			state = true
		case xmpp.StateGone:
			state = true
		case xmpp.ReceiptReceived:
			msgID = ex.ID
		case xmpp.MarkReceived:
			msgID = ex.ID
		case xmpp.MarkDisplayed:
			readed = true
			msgID = ex.ID
		}
	}
	if composing {
		return nil, nil
	}
	if state && msg.Body == "" {
		return nil, nil
	}
	if msgID != "" {
		id, err := strconv.ParseUint(msgID, 10, 64)
		if err != nil {
			return nil, err
		}
		msgType := o3.MSGDELIVERED
		if readed {
			msgType = o3.MSGREAD
		}
		drm, err := o3.NewDeliveryReceiptMessage(&a.Session, to, id, msgType)
		if err != nil {
			return nil, err
		}
		log.WithFields(map[string]interface{}{
			"tid":    to,
			"msg_id": id,
			"type":   msgType,
		}).Debug("update status of threema message")
		return drm, nil
	}

	// send text message
	msg3, err := o3.NewTextMessage(&a.Session, to, msg.Body)
	if err != nil {
		return nil, err
	}
	a.deliveredMSG[msg3.ID()] = msg.Id
	a.readedMSG[msg3.ID()] = msg.Id
	return msg3, nil
}
