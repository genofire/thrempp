package models

import (
	"regexp"

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

func ParseJID(jidString string) (jid *JID) {
	jidSplitTmp := jidRegex.FindAllStringSubmatch(jidString, -1)

	if len(jidSplitTmp) != 1 {
		return nil
	}
	jidSplit := jidSplitTmp[0]

	return &JID{
		Local:  jidSplit[1],
		Domain: jidSplit[2],
	}
}

func (jid *JID) String() string {
	if jid == nil {
		return ""
	}
	str := jid.Domain
	if str != "" && jid.Local != "" {
		str = jid.Local + "@" + str
	}
	return str
}

func (jid *JID) IsDomain() bool {
	return jid != nil && jid.Local == "" && jid.Domain != ""
}

var jidRegex *regexp.Regexp

func init() {
	jidRegex = regexp.MustCompile(`^(?:([^@/<>'\" ]+)@)?([^@/<>'\"]+)(?:/([^<>'\" ][^<>'\"]*))?$`)

	database.AddModel(&JID{})
}
