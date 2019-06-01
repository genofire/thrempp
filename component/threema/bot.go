package threema

import (
	"fmt"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"

	"dev.sum7.eu/genofire/golang-lib/database"

	"dev.sum7.eu/genofire/thrempp/models"
)

type Bot struct {
	jid     *models.JID
	threema *Threema
	server  o3.ThreemaRest
	logger  *log.Entry
}

func (t *Threema) getBot(jid *models.JID) *Bot {
	jidStr := jid.String()
	if bot, ok := t.bot[jidStr]; ok {
		return bot
	}
	if db := database.Read; db != nil && db.DB().Ping() == nil {
		if err := db.Where(jid).First(jid); err.RecordNotFound() {
			database.Write.Create(jid)
		} else if err != nil {
			log.Errorf("error getting jid %s from datatbase: %s", jid.String(), err.Error)
		}
	}
	bot := &Bot{
		jid:     jid,
		threema: t,
		server:  o3.ThreemaRest{},
		logger: log.WithFields(map[string]interface{}{
			"type": "threema",
			"jid":  jidStr,
		}),
	}
	t.bot[jidStr] = bot
	return bot
}
func (b *Bot) getAccount() (*Account, error) {
	return b.threema.getAccount(b.jid)
}

func (b *Bot) Handle(request string) string {
	switch request {
	case "generate":
		return b.cmdGenerate()
	case "help":
		return b.cmdHelp()
	}
	return fmt.Sprintf("command not found\n%s", b.cmdHelp())
}

func (b *Bot) cmdHelp() string {
	return `
	generate : generate  a threema id (if not exists)
		`
}

func (b *Bot) cmdGenerate() string {
	logger := b.logger.WithField("bot", "generate")
	// test if account already exists
	account, err := b.getAccount()
	if err == nil {
		return fmt.Sprintf("you already has the threema account with id: %s", string(account.TID))
	}

	// create account
	id, err := b.server.CreateIdentity()
	if err != nil {
		logger.Warnf("failed to generate: %s", err)
		return fmt.Sprintf("failed to create a threema account: %s", err)
	}

	// store account
	a := models.AccountThreema{}
	a.XMPPID = b.jid.ID
	a.TID = make([]byte, len(id.ID))
	a.LSK = make([]byte, len(id.LSK))
	copy(a.TID, id.ID[:])
	copy(a.LSK, id.LSK[:])
	database.Write.Create(&a)

	// fetch account and connect
	account, err = b.getAccount()
	if err != nil {
		logger.Warnf("failed to generate: %s", err)
	} else {
		if tid := string(account.TID); tid != "" {
			logger.WithField("threema", tid).Info("generate")
			return fmt.Sprintf("threema account with id: %s", tid)
		}
		logger.Warn("failed to generate")
	}
	return "failed to create a threema account"
}
