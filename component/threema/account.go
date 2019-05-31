package threema

import (
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
}

func (t *Threema) getAccount(jid *models.JID) *Account {
	if a, ok := t.accountJID[jid.String()]; ok {
		return a
	}
	account := models.AccountThreema{}

	database.Read.Where("xmpp_id = (?)",
		database.Read.Table(jid.TableName()).Select("id").Where(map[string]interface{}{
			"local":  jid.Local,
			"domain": jid.Domain,
		}).QueryExpr()).First(&account)

	var lsk [32]byte
	copy(lsk[:], account.LSK[:])
	tid, err := o3.NewThreemaID(string(account.TID), lsk, o3.AddressBook{})
	// TODO error handling
	if err != nil {
		return nil
	}
	tid.Nick = o3.NewPubNick("xmpp:" + jid.String())

	a := &Account{AccountThreema: account}
	a.XMPP = *jid
	a.Session = o3.NewSessionContext(tid)
	a.send, a.recieve, err = a.Session.Run()
	a.deliveredMSG = make(map[uint64]string)

	// TODO error handling
	if err != nil {
		return nil
	}

	go a.reciever(t.out)

	t.accountJID[jid.String()] = a
	return a
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
			out <- xMSG
		case o3.DeliveryReceiptMessage:
			if id, ok := a.deliveredMSG[msg.MsgID()]; ok {
				xMSG := xmpp.NewMessage("chat", msg.Sender().String(), a.XMPP.String(), "", "en")
				log.Warnf("found id %s", id)
				xMSG.Extensions = append(xMSG.Extensions, xmpp.ReceiptReceived{
					Id: id,
				})
				out <- xMSG
				delete(a.deliveredMSG, msg.MsgID())
			} else {
				log.Warnf("found not id in cache to announce received on xmpp side")
			}

		}
	}
}

func (a *Account) Send(to string, msg xmpp.Message) error {
	reci := ""
	for _, el := range msg.Extensions {
		switch ex := el.(type) {
		case *xmpp.ReceiptReceived:
			reci = ex.Id
		}
	}
	if reci != "" {
		id, err := strconv.ParseUint(reci, 10, 64)
		if err != nil {
			return err
		}
		drm, err := o3.NewDeliveryReceiptMessage(&a.Session, to, id, o3.MSGDELIVERED)
		if err != nil {
			return err
		}
		a.send <- drm
		log.WithFields(map[string]interface{}{
			"tid":    to,
			"msg_id": id,
		}).Debug("delivered")
		return nil
	}

	msg3, err := o3.NewTextMessage(&a.Session, to, msg.Body)
	if err != nil {
		return err
	}
	a.deliveredMSG[msg3.ID()] = msg.Id
	a.send <- msg3
	return nil
}
