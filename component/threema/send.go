package threema

import (
	"strconv"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp/stanza"
)

func (a *Account) Send(to string, msg stanza.Message) error {
	m, err := a.sending(to, msg)
	if err != nil {
		return err
	}
	if m != nil {
		a.send <- m
	}
	return nil
}
func (a *Account) sending(to string, msg stanza.Message) (o3.Message, error) {
	logger := log.WithFields(map[string]interface{}{
		"from": a.XMPP.String(),
		"to":   to,
	})

	chatState := false
	chatStateComposing := false

	msgStateID := ""
	msgStateRead := false

	for _, el := range msg.Extensions {
		switch ex := el.(type) {

		case *stanza.StateActive:
			chatState = true
		case *stanza.StateComposing:
			chatState = true
			chatStateComposing = true
		case *stanza.StateGone:
			chatState = true
		case *stanza.StateInactive:
			chatState = true
		case *stanza.StatePaused:
			chatState = true

		case *stanza.ReceiptReceived:
			msgStateID = ex.ID
		case *stanza.MarkReceived:
			msgStateID = ex.ID

		case *stanza.MarkDisplayed:
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
			logger.WithFields(map[string]interface{}{
				"msg_id": id,
				"type":   msgType,
			}).Debug("update status of threema message")
			return drm, nil
		}
		if chatState {
			tnm := o3.TypingNotificationMessage{}
			if chatStateComposing {
				tnm.OnOff = 0x1
			}
			logger.WithFields(map[string]interface{}{
				"state": chatStateComposing,
			}).Debug("not send typing")
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
	logger.WithFields(map[string]interface{}{
		"x_id": msg.Id,
		"t_id": msg3.ID(),
		"text": msg.Body,
	}).Debug("send text")
	return msg3, nil
}
