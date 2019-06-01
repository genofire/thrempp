package threema

import (
	"errors"
	"strconv"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"
	"gosrc.io/xmpp"

	"dev.sum7.eu/genofire/golang-lib/database"

	"dev.sum7.eu/genofire/thrempp/models"
)

type Account struct {
	models.AccountThreema
	Session      o3.SessionContext
	send         chan<- o3.Message
	recieve      <-chan o3.ReceivedMsg
	deliveredMSG map[uint64]string
	readedMSG    map[uint64]string
}

func (t *Threema) getAccount(jid *models.JID) (*Account, error) {
	if a, ok := t.accountJID[jid.String()]; ok {
		return a, nil
	}
	account := models.AccountThreema{}

	if database.Read == nil {
		return nil, errors.New("no database connection")
	}

	database.Read.Where("xmpp_id = (?)",
		database.Read.Table(jid.TableName()).Select("id").Where(map[string]interface{}{
			"local":  jid.Local,
			"domain": jid.Domain,
		}).QueryExpr()).First(&account)

	var lsk [32]byte
	copy(lsk[:], account.LSK[:])
	tid, err := o3.NewThreemaID(string(account.TID), lsk, o3.AddressBook{})
	if err != nil {
		return nil, err
	}
	tid.Nick = o3.NewPubNick("xmpp:" + jid.String())

	a := &Account{AccountThreema: account}
	a.Session = o3.NewSessionContext(tid)
	a.send, a.recieve, err = a.Session.Run()

	if err != nil {
		return nil, err
	}

	a.XMPP = *jid
	a.deliveredMSG = make(map[uint64]string)
	a.readedMSG = make(map[uint64]string)

	go a.reciever(t.out)

	t.accountJID[jid.String()] = a
	return a, nil
}

func (a *Account) reciever(out chan<- xmpp.Packet) {
	for receivedMessage := range a.recieve {
		if receivedMessage.Err != nil {
			log.Warnf("Error Receiving Message: %s\n", receivedMessage.Err)
			continue
		}
		switch msg := receivedMessage.Msg.(type) {
		case o3.TextMessage:
			sender := msg.Sender().String()
			if string(a.TID) == sender {
				continue
			}
			xMSG := xmpp.NewMessage("chat", sender, a.XMPP.String(), strconv.FormatUint(msg.ID(), 10), "en")
			xMSG.Body = msg.Text()
			xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptRequest{})
			xMSG.Extensions = append(xMSG.Extensions, xmpp.ChatMarkerMarkable{})
			out <- xMSG

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
				out <- xMSG
			}
		}
	}
}

func (a *Account) Send(to string, msg xmpp.Message) error {
	msgID := ""
	readed := false
	for _, el := range msg.Extensions {
		switch ex := el.(type) {
		case *xmpp.ReceiptReceived:
			msgID = ex.Id
		case *xmpp.ChatMarkerReceived:
			msgID = ex.Id
		case *xmpp.ChatMarkerDisplayed:
			readed = true
			msgID = ex.Id
		}
	}
	if msgID != "" {
		id, err := strconv.ParseUint(msgID, 10, 64)
		if err != nil {
			return err
		}
		msgType := o3.MSGDELIVERED
		if readed {
			msgType = o3.MSGREAD
		}
		drm, err := o3.NewDeliveryReceiptMessage(&a.Session, to, id, msgType)
		if err != nil {
			return err
		}
		a.send <- drm
		log.WithFields(map[string]interface{}{
			"tid":    to,
			"msg_id": id,
			"type":   msgType,
		}).Debug("update status of threema message")
		return nil
	}

	msg3, err := o3.NewTextMessage(&a.Session, to, msg.Body)
	if err != nil {
		return err
	}
	a.deliveredMSG[msg3.ID()] = msg.Id
	a.readedMSG[msg3.ID()] = msg.Id
	a.send <- msg3
	return nil
}
