package models

import (
	"github.com/jinzhu/gorm"

	"dev.sum7.eu/genofire/golang-lib/database"
)

type AccountThreema struct {
	gorm.Model
	XMPPID uint
	XMPP   JID
	TID    []byte
	LSK    []byte
}

func init() {
	database.AddModel(&AccountThreema{})
}
