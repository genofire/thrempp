package threema

import (
	"strconv"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp"
)

func (a *Account) receiver(out chan<- xmpp.Packet) {
	for receivedMessage := range a.receive {
		if p := a.receiving(receivedMessage); p != nil {
			out <- p
		}
	}
}
func (a *Account) receiving(receivedMessage o3.ReceivedMsg) xmpp.Packet {
	if receivedMessage.Err != nil {
		log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
		return nil
	}
	switch msg := receivedMessage.Msg.(type) {
	case o3.TextMessage:
		sender := msg.Sender().String()
		if string(a.TID) == sender {
			return nil
		}
		xMSG := xmpp.NewMessage("chat", sender, a.XMPP.String(), strconv.FormatUint(msg.ID(), 10), "en")
		xMSG.Body = msg.Text()
		xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptRequest{})
		xMSG.Extensions = append(xMSG.Extensions, xmpp.ChatMarkerMarkable{})
		return xMSG

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
			return xMSG
		}
	}
	return nil
}
