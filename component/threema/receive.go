package threema

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp"
)

func (a *Account) receiver(out chan<- xmpp.Packet) {
	for receivedMessage := range a.receive {
		if receivedMessage.Err != nil {
			log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
			xMSG := xmpp.NewMessage("chat", "", a.XMPP.String(), "", "en")
			xMSG.Body = fmt.Sprintf("error on decoding message:\n%v", receivedMessage.Err)
			out <- xMSG
			continue
		}
		sender := receivedMessage.Msg.Sender().String()
		if string(a.TID) == sender {
			continue
		}
		if p, err := a.receiving(receivedMessage.Msg); err != nil {
			xMSG := xmpp.NewMessage("chat", sender, a.XMPP.String(), "", "en")
			xMSG.Body = fmt.Sprintf("error on decoding message: %s\n%v", err, receivedMessage.Msg.Serialize())
			out <- xMSG
		} else if p != nil {
			out <- p
		}
	}
}

func requestExtensions(xMSG *xmpp.Message) {
	xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptRequest{})
	xMSG.Extensions = append(xMSG.Extensions, xmpp.Markable{})
	xMSG.Extensions = append(xMSG.Extensions, xmpp.StateActive{})
}

func (a *Account) receiving(receivedMessage o3.Message) (xmpp.Packet, error) {
	logger := log.WithFields(map[string]interface{}{
		"from": receivedMessage.Sender().String(),
		"to":   a.XMPP.String(),
	})
	switch msg := receivedMessage.(type) {
	case o3.TextMessage:
		sender := msg.Sender().String()
		xMSG := xmpp.NewMessage("chat", sender, a.XMPP.String(), strconv.FormatUint(msg.ID(), 10), "en")
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
		xMSG := xmpp.NewMessage("chat", msg.Sender().String(), a.XMPP.String(), "", "en")
		state := ""

		if msg.Status() == o3.MSGDELIVERED {
			state = "delivered"
			if id, ok := a.deliveredMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptReceived{ID: id})
				xMSG.Extensions = append(xMSG.Extensions, xmpp.MarkReceived{ID: id})
				delete(a.deliveredMSG, msgID)
			} else {
				logger.Warnf("found not id in cache to announce received on xmpp side")
			}
		}
		if msg.Status() == o3.MSGREAD {
			state = "displayed"
			if id, ok := a.readedMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, xmpp.MarkDisplayed{ID: id})
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
		xMSG := xmpp.NewMessage("chat", msg.Sender().String(), a.XMPP.String(), strconv.FormatUint(msg.ID(), 10), "en")
		if msg.OnOff != 0 {
			logger.Debug("composing")
			xMSG.Extensions = append(xMSG.Extensions, xmpp.StateComposing{})
		} else {
			logger.Debug("inactive")
			xMSG.Extensions = append(xMSG.Extensions, xmpp.StateInactive{})
		}
		return xMSG, nil
	}
	return nil, errors.New("not known data format")
}
