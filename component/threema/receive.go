package threema

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func (a *Account) receiver() {
	for receivedMessage := range a.receive {
		if receivedMessage.Err != nil {
			log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
			xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, To: a.XMPP.String()})
			xMSG.Body = fmt.Sprintf("error on decoding message:\n%v", receivedMessage.Err)
			a.xmpp <- xMSG
			continue
		}
		header := receivedMessage.Msg.Header()
		sender := header.Sender.String()
		if string(a.TID) == sender {
			continue
		}
		if p, gh, err := a.receiving(receivedMessage.Msg); err != nil {
			xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String()})
			xMSG.Body = fmt.Sprintf("error on decoding message: %s\n%v", err, receivedMessage.Msg)
			a.xmpp <- xMSG
		} else {
			if gh != nil {
				xid := &xmpp.Jid{
					Node:   a.XMPP.Local,
					Domain: a.XMPP.Domain,
				}
				id := strFromThreemaGroup(gh)
				if len(a.XMPPResource[id]) == 0 {
					xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String()})
					//TODO please join
					xMSG.Body = fmt.Sprintf(`ERROR: group message not delievered, please join
xmpp:%s@{{DOMAIN}}?join`, id)
					a.xmpp <- xMSG
					continue
				}
				for r := range a.XMPPResource[id] {
					xid.Resource = r
					switch m := p.(type) {
					case stanza.Message:
						m.Attrs.To = xid.Full()
						a.xmpp <- m
					}
				}
			} else {
				a.xmpp <- p
			}
		}
	}
}

func requestExtensions(xMSG *stanza.Message) {
	xMSG.Extensions = append(xMSG.Extensions, stanza.ReceiptRequest{})
	xMSG.Extensions = append(xMSG.Extensions, stanza.Markable{})
	xMSG.Extensions = append(xMSG.Extensions, stanza.StateActive{})
}

func (a *Account) receiving(receivedMessage o3.Message) (stanza.Packet, *o3.GroupMessageHeader, error) {
	header := receivedMessage.Header()
	sender := header.Sender.String()
	logger := log.WithFields(map[string]interface{}{
		"from": sender,
		"to_t": header.Recipient.String(),
		"to":   a.XMPP.String(),
	})
	sender = strings.ToLower(sender)
	switch msg := receivedMessage.(type) {
	case *o3.TextMessage:
		dbText := "recv text"
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String(), Id: strconv.FormatUint(header.ID, 10)})
		if msg.GroupMessageHeader != nil {
			to := a.XMPP.String()
			ad := strings.SplitN(msg.Body, "=", 2)
			from := sender
			if len(ad) == 2 {
				from = strings.ToLower(ad[0])[:len(ad[0])-3]
			}
			xMSG = stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeGroupchat, From: jidFromThreemaGroup(from, msg.GroupMessageHeader), To: to, Id: strconv.FormatUint(header.ID, 10)})
			if len(ad) == 2 {
				xMSG.Body = ad[1][4:]
			} else {
				xMSG.Body = msg.Body
			}
			dbText = "recv grouptext"
		} else {
			xMSG.Body = msg.Body
			requestExtensions(&xMSG)
		}
		logger.WithFields(map[string]interface{}{
			"from_x": xMSG.From,
			"id":     xMSG.Id,
			"text":   xMSG.Body,
		}).Debug(dbText)
		return xMSG, msg.GroupMessageHeader, nil
	case *o3.AudioMessage:
		if a.threema.httpUploadPath == "" {
			return nil, nil, errors.New("no place to store files at transport configurated")
		}
		data, err := msg.GetData()
		if err != nil {
			logger.Warnf("unable to read data from message: %s", err)
			return nil, nil, err
		}
		xMSG, err := a.FileToXMPP(sender, header.ID, "mp3", data)
		if err != nil {
			logger.Warnf("unable to create data from message: %s", err)
			return nil, nil, err
		}
		requestExtensions(&xMSG)
		logger.WithField("url", xMSG.Body).Debug("recv audio")
		return xMSG, nil, nil

	case *o3.ImageMessage:
		if a.threema.httpUploadPath == "" {
			return nil, nil, errors.New("no place to store files at transport configurated")
		}
		data, err := msg.GetData(a.ThreemaID)
		if err != nil {
			logger.Warnf("unable to read data from message: %s", err)
			return nil, nil, err
		}
		xMSG, err := a.FileToXMPP(sender, header.ID, "jpg", data)
		if err != nil {
			logger.Warnf("unable to create data from message: %s", err)
			return nil, nil, err
		}
		requestExtensions(&xMSG)
		logger.WithField("url", xMSG.Body).Debug("recv image")
		return xMSG, nil, nil

	case *o3.DeliveryReceiptMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String()})
		state := ""

		if msg.Status == o3.MSGDELIVERED {
			state = "delivered"
			if id, ok := a.deliveredMSG[msg.MessageID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, stanza.ReceiptReceived{ID: id})
				xMSG.Extensions = append(xMSG.Extensions, stanza.MarkReceived{ID: id})
				delete(a.deliveredMSG, msg.MessageID)
			} else {
				logger.Warnf("found not id in cache to announce received on xmpp side")
			}
		}
		if msg.Status == o3.MSGREAD {
			state = "displayed"
			if id, ok := a.readedMSG[msg.MessageID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, stanza.MarkDisplayed{ID: id})
				delete(a.readedMSG, msg.MessageID)
			} else {
				logger.Warnf("found not id in cache to announce readed on xmpp side")
			}
		}

		if len(xMSG.Extensions) > 0 {
			logger.WithField("state", state).Debug("recv state")
			return xMSG, nil, nil
		}
		return nil, nil, nil
	case *o3.TypingNotificationMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String(), Id: strconv.FormatUint(header.ID, 10)})
		if msg.OnOff != 0 {
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateComposing{})
		} else {
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateInactive{})
		}
		logger.WithField("on", msg.OnOff).Debug("recv typing")
		return xMSG, nil, nil
	}
	return nil, nil, errors.New("not known data format")
}
