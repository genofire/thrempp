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
	Session o3.SessionContext
	send    chan<- o3.Message
	recieve <-chan o3.ReceivedMsg
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

	// TODO error handling
	if err != nil {
		return nil
	}

	go a.reciever(t.out)

	t.accountJID[jid.String()] = a
	t.accountTID[string(a.TID)] = a
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
			out <- xMSG
		case o3.DeliveryReceiptMessage:
			// msg.MsgID()

		}
	}
}

func (a *Account) Send(to string, msg string) error {
	return a.Session.SendTextMessage(to, msg, a.send)
}
