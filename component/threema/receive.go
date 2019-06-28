package threema

import (
	"errors"
	"fmt"
	"strconv"

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
		sender := receivedMessage.Msg.Sender().String()
		if string(a.TID) == sender {
			continue
		}
		if p, err := a.receiving(receivedMessage.Msg); err != nil {
			xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String()})
			xMSG.Body = fmt.Sprintf("error on decoding message: %s\n%v", err, receivedMessage.Msg.Serialize())
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

func (a *Account) receiving(receivedMessage o3.Message) (stanza.Packet, error) {
	logger := log.WithFields(map[string]interface{}{
		"from": receivedMessage.Sender().String(),
		"to":   a.XMPP.String(),
	})
	switch msg := receivedMessage.(type) {
	case o3.TextMessage:
		sender := msg.Sender().String()
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: sender, To: a.XMPP.String(), Id: strconv.FormatUint(msg.ID(), 10)})
		xMSG.Body = msg.Text()
		requestExtensions(&xMSG)
		logger.WithField("text", xMSG.Body).Debug("send text")
		return xMSG, nil

	case o3.AudioMessage:
		if a.threema.httpUploadPath == "" {
			return nil, errors.New("no place to store files at transport configurated")
		}
		data, err := msg.GetAudioData(a.Session)
		if err != nil {
			logger.Warnf("unable to read data from message: %s", err)
			return nil, err
		}
		xMSG, err := a.FileToXMPP(msg.Sender().String(), msg.ID(), "mp3", data)
		if err != nil {
			logger.Warnf("unable to create data from message: %s", err)
			return nil, err
		}
		xMSG.Type = "chat"
		requestExtensions(&xMSG)
		logger.WithField("url", xMSG.Body).Debug("send audio")
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
		xMSG, err := a.FileToXMPP(msg.Sender().String(), msg.ID(), "jpg", data)
		if err != nil {
			logger.Warnf("unable to create data from message: %s", err)
			return nil, err
		}
		xMSG.Type = "chat"
		requestExtensions(&xMSG)
		logger.WithField("url", xMSG.Body).Debug("send image")
		return xMSG, nil

	case o3.DeliveryReceiptMessage:
		msgID := msg.MsgID()
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: msg.Sender().String(), To: a.XMPP.String()})
		state := ""

		if msg.Status() == o3.MSGDELIVERED {
			state = "delivered"
			if id, ok := a.deliveredMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, stanza.ReceiptReceived{ID: id})
				xMSG.Extensions = append(xMSG.Extensions, stanza.MarkReceived{ID: id})
				delete(a.deliveredMSG, msgID)
			} else {
				logger.Warnf("found not id in cache to announce received on xmpp side")
			}
		}
		if msg.Status() == o3.MSGREAD {
			state = "displayed"
			if id, ok := a.readedMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, stanza.MarkDisplayed{ID: id})
				delete(a.readedMSG, msgID)
			} else {
				logger.Warnf("found not id in cache to announce readed on xmpp side")
			}
		}

		if len(xMSG.Extensions) > 0 {
			logger.WithField("state", state).Debug("send state")
			return xMSG, nil
		}
		return nil, nil
	case o3.TypingNotificationMessage:
		xMSG := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: msg.Sender().String(), To: a.XMPP.String(), Id: strconv.FormatUint(msg.ID(), 10)})
		if msg.OnOff != 0 {
			logger.Debug("composing")
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateComposing{})
		} else {
			logger.Debug("inactive")
			xMSG.Extensions = append(xMSG.Extensions, stanza.StateInactive{})
		}
		return xMSG, nil
	}
	return nil, errors.New("not known data format")
}
