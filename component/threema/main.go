package threema

import (
	"errors"
	"strings"

	"github.com/bdlm/log"
	"gosrc.io/xmpp"

	"dev.sum7.eu/genofire/golang-lib/database"

	"dev.sum7.eu/genofire/thrempp/component"
	"dev.sum7.eu/genofire/thrempp/models"
)

type Threema struct {
	component.Component
	out            chan xmpp.Packet
	accountJID     map[string]*Account
	bot            map[string]*Bot
	httpUploadPath string
	httpUploadURL  string
}

func NewThreema(config map[string]interface{}) (component.Component, error) {
	t := &Threema{
		out:        make(chan xmpp.Packet),
		accountJID: make(map[string]*Account),
		bot:        make(map[string]*Bot),
	}
	if pathI, ok := config["http_upload_path"]; ok {
		if path, ok := pathI.(string); ok {
			t.httpUploadPath = path
		} else {
			return nil, errors.New("wrong format of http_upload_path")
		}
	}
	if urlI, ok := config["http_upload_url"]; ok {
		if url, ok := urlI.(string); ok {
			t.httpUploadURL = url
		} else {
			return nil, errors.New("wrong format of http_upload_url")
		}
	}
	return t, nil
}

func (t *Threema) Connect() (chan xmpp.Packet, error) {
	var jids []*models.JID
	database.Read.Find(&jids)
	for _, jid := range jids {
		logger := log.WithField("jid", jid.String())
		a, err := t.getAccount(jid)
		if err != nil {
			logger.Warnf("unable to connect%s", err)
			continue
		}
		logger = logger.WithField("threema", string(a.TID))
		logger.Info("connected")
	}
	return t.out, nil
}
func (t *Threema) Send(packet xmpp.Packet) {
	if p := t.send(packet); p != nil {
		t.out <- p
	}
}
func (t *Threema) send(packet xmpp.Packet) xmpp.Packet {
	switch p := packet.(type) {
	case xmpp.Message:
		from := models.ParseJID(p.PacketAttrs.From)
		to := models.ParseJID(p.PacketAttrs.To)

		if to.IsDomain() {
			if from == nil {
				log.Warn("receive message without sender")
				return nil
			}
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = t.getBot(from).Handle(p.Body)
			return msg
		}

		account, err := t.getAccount(from)
		if err != nil {
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = "It was not possible to send, because we have no account for you.\nPlease generate one, by sending `generate` to this gateway"
			return msg
		}

		threemaID := strings.ToUpper(to.Local)
		if err := account.Send(threemaID, p); err != nil {
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = err.Error()
			return msg
		}
	default:
		log.Warnf("unknown package: %v", p)
	}
	return nil
}

func init() {
	component.AddComponent("threema", NewThreema)
}
