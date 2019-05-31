package threema

import (
	"strings"

	"github.com/bdlm/log"
	"gosrc.io/xmpp"

	"dev.sum7.eu/genofire/thrempp/component"
	"dev.sum7.eu/genofire/thrempp/models"
)

type Threema struct {
	component.Component
	out        chan xmpp.Packet
	accountJID map[string]*Account
	accountTID map[string]*Account
}

func NewThreema(config map[string]interface{}) (component.Component, error) {
	t := &Threema{
		out:        make(chan xmpp.Packet),
		accountJID: make(map[string]*Account),
		accountTID: make(map[string]*Account),
	}
	// TODO load accounts on startup
	return t, nil
}

func (t *Threema) Connect() (chan xmpp.Packet, error) {
	return t.out, nil
}
func (t *Threema) Send(packet xmpp.Packet) {
	switch p := packet.(type) {
	case xmpp.Message:
		from := models.ParseJID(p.PacketAttrs.From)
		to := models.ParseJID(p.PacketAttrs.To)

		logger := log.WithFields(map[string]interface{}{
			"from": from,
			"to":   to,
		})
		logger.Debug(p.Body)
		if to.IsDomain() {
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = t.Bot(from, p.Body)
			t.out <- msg
			return
		}

		account := t.getAccount(from)
		if account == nil {
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = "It was not possible to send, becouse we have no account for you.\nPlease generate one, by sending `generate` to this gateway"
			t.out <- msg
			return
		}

		threemaID := strings.ToUpper(to.Local)
		if err := account.Send(threemaID, p.Body); err != nil {
			msg := xmpp.NewMessage("chat", "", from.String(), "", "en")
			msg.Body = err.Error()
			t.out <- msg
		}
	default:
		log.Warnf("unkown package%v", p)
	}
}

func init() {
	component.AddComponent("threema", NewThreema)
}
