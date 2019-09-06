package threema

import (
	"encoding/base32"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp/stanza"
)

func (a *Account) receiver(out chan<- stanza.Packet) {
	for receivedMessage := range a.receive {
		if receivedMessage.Err != nil {
			log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
			xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, To: a.XMPP.String()})
			xMSG.Body = fmt.Sprintf("error on decoding message:\n%v", receivedMessage.Err)
			out <- xMSG
			continue
		}
		header := receivedMessage.Msg.Header()
		sender := header.Sender.String()
		if string(a.TID) == sender {
			continue
		}
		if p, err := a.receiving(receivedMessage.Msg); err != nil {
			xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String()})
			xMSG.Body = fmt.Sprintf("error on decoding message: %s\n%v", err, receivedMessage.Msg)
			out <- xMSG
		} else if p != nil {
			out <- p
		}
	}
}

func requestExtensions(xMSG *stanza.Message) {
	xMSG.Extensions = append(xMSG.Extensions, stanza.ReceiptRequest{})
	xMSG.Extensions = append(xMSG.Extensions, stanza.Markable{})
	xMSG.Extensions = append(xMSG.Extensions, stanza.StateActive{})
}

func jidFromThreemaGroup(sender string, header *o3.GroupMessageHeader) string {
	cid := strings.ToLower(header.CreatorID.String())
	gid := strings.ToLower(base32.StdEncoding.EncodeToString(header.GroupID[:]))
	return fmt.Sprintf("%s-%s@{{DOMAIN}}/%s", cid, gid, sender)
}

func (a *Account) receiving(receivedMessage o3.Message) (stanza.Packet, error) {
	header := receivedMessage.Header()
	sender := header.Sender.String()
	logger := log.WithFields(map[string]interface{}{
		"from": sender,
		"to_t": header.Recipient.String(),
		"to":   a.XMPP.String(),
	})
	sender = strings.ToLower(sender)
	switch msg := receivedMessage.(type) {
	case *o3.GroupTextMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeGroupchat, From: jidFromThreemaGroup(sender, msg.GroupMessageHeader), To: a.XMPP.String(), Id: strconv.FormatUint(header.ID, 10)})
		xMSG.Body = msg.Body
		requestExtensions(&xMSG)
		logger.WithFields(map[string]interface{}{
			"from_x": xMSG.From,
			"id":     xMSG.Id,
			"text":   xMSG.Body,
		}).Debug("recv grouptext")
		return xMSG, nil
	case *o3.TextMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String(), Id: strconv.FormatUint(header.ID, 10)})
		xMSG.Body = msg.Body
		requestExtensions(&xMSG)
		logger.WithFields(map[string]interface{}{
			"from_x": xMSG.From,
			"id":     xMSG.Id,
			"text":   xMSG.Body,
		}).Debug("recv text")
		return xMSG, nil
	/*
		case o3.AudioMessage:
			if a.threema.httpUploadPath == "" {
				return nil, errors.New("no place to store files at transport configurated")
			}
			data, err := msg.GetAudioData(a.Session)
			if err != nil {
				logger.Warnf("unable to read data from message: %s", err)
				return nil, err
			}
			xMSG, err := a.FileToXMPP(sender.String(), header.ID, "mp3", data)
			if err != nil {
				logger.Warnf("unable to create data from message: %s", err)
				return nil, err
			}
			xMSG.Type = "chat"
			requestExtensions(&xMSG)
			logger.WithField("url", xMSG.Body).Debug("recv audio")
			return xMSG, nil

		case o3.ImageMessage:
			if a.threema.httpUploadPath == "" {
				return nil, errors.New("no place to store files at transport configurated")
			}
			data, err := msg.GetImageData(a.Session)
			if err != nil {
				logger.Warnf("unable to read data from message: %s", err)
				return nil, err
			}
			xMSG, err := a.FileToXMPP(sender.String(), header.ID, "jpg", data)
			if err != nil {
				logger.Warnf("unable to create data from message: %s", err)
				return nil, err
			}
			xMSG.Type = "chat"
			requestExtensions(&xMSG)
			logger.WithField("url", xMSG.Body).Debug("recv image")
			return xMSG, nil
	*/
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
			return xMSG, nil
		}
		return nil, nil
	case *o3.TypingNotificationMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String(), Id: strconv.FormatUint(header.ID, 10)})
		if msg.OnOff != 0 {
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateComposing{})
		} else {
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateInactive{})
		}
		logger.WithField("on", msg.OnOff).Debug("recv typing")
		return xMSG, nil
	}
	return nil, errors.New("not known data format")
}
