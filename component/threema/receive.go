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
		sender := receivedMessage.Msg.Sender().String()
		if string(a.TID) == sender {
			continue
		}
		if p, err := a.receiving(receivedMessage); err != nil {
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
	xMSG.Extensions = append(xMSG.Extensions, xmpp.ChatMarkerMarkable{})
}

func (a *Account) receiving(receivedMessage o3.ReceivedMsg) (xmpp.Packet, error) {
	if receivedMessage.Err != nil {
		log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
		return nil, receivedMessage.Err
	}
	switch msg := receivedMessage.Msg.(type) {
	case o3.TextMessage:
		sender := msg.Sender().String()
		xMSG := xmpp.NewMessage("chat", sender, a.XMPP.String(), strconv.FormatUint(msg.ID(), 10), "en")
		xMSG.Body = msg.Text()
		requestExtensions(&xMSG)
		return xMSG, nil

	case o3.ImageMessage:
		if a.threema.httpUploadPath == "" {
			return nil, errors.New("no place to store files at transport configurated")
		}
		data, err := msg.GetImageData(a.Session)
		if err != nil {
			log.Warnf("unable to read data from message: %s", err)
			return nil, err
		}
		xMSG, err := a.FileToXMPP(msg.Sender().String(), msg.ID(), "jpg", data)
		if err != nil {
			log.Warnf("unable to create data from message: %s", err)
			return nil, err
		}
		xMSG.Type = "chat"
		requestExtensions(&xMSG)
		return xMSG, nil

	case o3.DeliveryReceiptMessage:
		msgID := msg.MsgID()
		xMSG := xmpp.NewMessage("chat", msg.Sender().String(), a.XMPP.String(), "", "en")

		if msg.Status() == o3.MSGDELIVERED {
			if id, ok := a.deliveredMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptReceived{Id: id})
				xMSG.Extensions = append(xMSG.Extensions, xmpp.ChatMarkerReceived{Id: id})
				delete(a.deliveredMSG, msgID)
			} else {
				log.Warnf("found not id in cache to announce received on xmpp side")
			}
		}
		if msg.Status() == o3.MSGREAD {
			if id, ok := a.readedMSG[msgID]; ok {
				xMSG.Extensions = append(xMSG.Extensions, xmpp.ChatMarkerDisplayed{Id: id})
				delete(a.readedMSG, msgID)
			} else {
				log.Warnf("found not id in cache to announce readed on xmpp side")
			}
		}

		if len(xMSG.Extensions) > 0 {
			return xMSG, nil
		}
		return nil, nil
	}
	return nil, errors.New("not known data format")
}
