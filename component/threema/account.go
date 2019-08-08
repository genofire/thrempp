package threema

import (
	"errors"

	"github.com/o3ma/o3"

	"dev.sum7.eu/genofire/golang-lib/database"

	"dev.sum7.eu/sum7/thrempp/models"
)

type Account struct {
	models.AccountThreema
	threema      *Threema
	Session      o3.SessionContext
	send         chan<- o3.Message
	receive      <-chan o3.ReceivedMsg
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

	a := &Account{
		AccountThreema: account,
		Session:        o3.NewSessionContext(tid),
		threema:        t,
	}
	a.send, a.receive, err = a.Session.Run()

	if err != nil {
		return nil, err
	}

	a.XMPP = *jid
	a.deliveredMSG = make(map[uint64]string)
	a.readedMSG = make(map[uint64]string)

	go a.receiver(t.out)

	t.accountJID[jid.String()] = a
	return a, nil
}
