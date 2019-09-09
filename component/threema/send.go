package threema

import (
	"encoding/base32"
	"strconv"
	"strings"

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
	from := string(a.AccountThreema.TID)
	logger := log.WithFields(map[string]interface{}{
		"from":   a.XMPP.String(),
		"from_t": from,
		"to":     to,
	})
	msg3ID := o3.NewMsgID()
	header := &o3.MessageHeader{
		Sender:    o3.NewIDString(from),
		ID:        msg3ID,
		Recipient: o3.NewIDString(to),
		PubNick:   a.ThreemaID.Nick,
	}
	var groupHeader *o3.GroupMessageHeader
	if msg.Type == stanza.MessageTypeGroupchat {
		toA := strings.SplitN(to, "-", 2)
		gid, err := base32.StdEncoding.DecodeString(toA[1])
		if err != nil {
			return nil, err
		}
		groupHeader = &o3.GroupMessageHeader{
			CreatorID: o3.NewIDString(toA[0]),
		}
		copy(groupHeader.GroupID[:], gid)
	}
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
			drm := &o3.DeliveryReceiptMessage{
				MessageHeader: header,
				Status:        o3.MSGDELIVERED,
				MessageID:     id,
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
			tnm := &o3.TypingNotificationMessage{
				MessageHeader: header,
			}
			if chatStateComposing {
				tnm.OnOff = 0x1
			}
			logger.WithFields(map[string]interface{}{
				"on": tnm.OnOff,
			}).Debug("send typing")
			return tnm, nil
		}
	}

	// send text message
	msg3 := &o3.TextMessage{
		GroupMessageHeader: groupHeader,
		MessageHeader:      header,
		Body:               msg.Body,
	}
	logger = logger.WithFields(map[string]interface{}{
		"x_id": msg.Id,
		"t_id": msg3ID,
		"text": msg.Body,
	})
	if groupHeader != nil {
		logger.Debug("send grouptext")
		// TODO iterate of all occupants
		return msg3, nil
	}
	a.deliveredMSG[msg3ID] = msg.Id
	a.readedMSG[msg3ID] = msg.Id
	logger.Debug("send text")
	return msg3, nil
}
