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
	msg3To := o3.NewIDString(to)
	msg3From := o3.NewIDString(string(a.AccountThreema.TID))

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
			drm := o3.DeliveryReceiptMessage{
				MessageHeader: &o3.MessageHeader{
					Sender:    msg3From,
					ID:        id,
					Recipient: msg3To,
				},
				Status: o3.MSGDELIVERED,
			}
			if msgStateRead {
				drm.Status = o3.MSGREAD
			}
			logger.WithFields(map[string]interface{}{
				"msg_id": id,
				"type":   drm.Status,
			}).Debug("update status of threema message")
			return drm, nil
		}
		if chatState {
			tnm := o3.TypingNotificationMessage{
				MessageHeader: &o3.MessageHeader{
					Sender: o3.NewIDString(string(a.AccountThreema.TID)),
				},
			}
			if chatStateComposing {
				tnm.OnOff = 0x1
			}
			logger.WithFields(map[string]interface{}{
				"state": chatStateComposing,
			}).Debug("not send typing")
			return nil, nil
		}
	}
	msg3ID := o3.NewMsgID()

	// send text message
	msg3 := o3.TextMessage{
		MessageHeader: &o3.MessageHeader{
			Sender:    o3.NewIDString(string(a.AccountThreema.TID)),
			ID:        msg3ID,
			Recipient: msg3To,
		},
		Body: msg.Body,
	}
	a.deliveredMSG[msg3ID] = msg.Id
	a.readedMSG[msg3ID] = msg.Id
	logger.WithFields(map[string]interface{}{
		"x_id": msg.Id,
		"t_id": msg3ID,
		"text": msg.Body,
	}).Debug("send text")
	return msg3, nil
}
