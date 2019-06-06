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
	chatState := false
	chatStateComposing := false
	msgStateID := ""
	msgStateRead := false

	for _, el := range msg.Extensions {
		switch ex := el.(type) {
		case xmpp.StateActive:
			chatState = true
		case xmpp.StateComposing:
			chatState = true
			chatStateComposing = true
		case xmpp.StateGone:
			chatState = true
		case xmpp.StateInactive:
			chatState = true
		case xmpp.StatePaused:
			chatState = true
		case xmpp.ReceiptReceived:
			msgStateID = ex.ID
		case xmpp.MarkReceived:
			msgStateID = ex.ID
		case xmpp.MarkDisplayed:
			msgStateRead = true
			msgStateID = ex.ID
		}
	}
	if msg.Body == "" {
		if msgStateID != "" {
			id, err := strconv.ParseUint(msgStateID, 10, 64)
			if err != nil {
				return nil, err
			}
			msgType := o3.MSGDELIVERED
			if msgStateRead {
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
		if chatStateComposing {
			return nil, nil
		}
		if chatState {
			return nil, nil
		}
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
