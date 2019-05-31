package models

import (
	"github.com/jinzhu/gorm"

	"dev.sum7.eu/genofire/golang-lib/database"
)

type JID struct {
	gorm.Model
	Local  string
	Domain string
}

func (j *JID) TableName() string {
	return "jid"
}

func init() {
	database.AddModel(&JID{})
}
