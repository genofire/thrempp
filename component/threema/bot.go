package threema

import (
	"fmt"

	"github.com/bdlm/log"
	"github.com/o3ma/o3"

	"dev.sum7.eu/genofire/golang-lib/database"

	"dev.sum7.eu/genofire/thrempp/models"
)

func (t *Threema) Bot(from *models.JID, request string) string {
	server := o3.ThreemaRest{}
	logger := log.WithFields(map[string]interface{}{
		"type": "threema",
		"jid":  from.String(),
	})

	switch request {
	case "generate":

		// test if account already exists
		account := t.getAccount(from)
		if account != nil {
			return fmt.Sprintf("you already has the threema account with id: %s", string(account.TID))
		}

		// create account
		id, err := server.CreateIdentity()
		if err != nil {
			logger.Warnf("failed to generate: %s", err)
			return fmt.Sprintf("failed to create a threema account: %s", err)
		}
		//TODO works it
		if err := database.Read.Where(from).First(from); err != nil {
			database.Write.Create(from)
		}

		// store account
		a := models.AccountThreema{}
		a.XMPPID = from.ID
		a.TID = make([]byte, len(id.ID))
		a.LSK = make([]byte, len(id.LSK))
		copy(a.TID, id.ID[:])
		copy(a.LSK, id.LSK[:])
		database.Write.Create(&a)

		// fetch account and connect
		account = t.getAccount(from)
		tid := string(account.TID)
		if tid != "" {
			logger.WithField("threema", tid).Info("generate")
			return fmt.Sprintf("threema account with id: %s", tid)
		}
		logger.Warn("failed to generate")
		return "failed to create a threema account"
	}
	return "command not supported"
}
