package threema

import (
	"strings"

	"github.com/bdlm/log"
	"gosrc.io/xmpp"

	"dev.sum7.eu/genofire/thrempp/component"
)

type Threema struct {
	component.Component
	out chan xmpp.Packet
}

func NewThreema(config map[string]interface{}) (component.Component, error) {
	return &Threema{}, nil
}

func (t *Threema) Connect() (chan xmpp.Packet, error) {
	t.out = make(chan xmpp.Packet)
	return t.out, nil
}
func (t *Threema) Send(packet xmpp.Packet) {
	switch p := packet.(type) {
	case xmpp.Message:
		attrs := p.PacketAttrs
		account := t.getAccount(attrs.From)
		log.WithFields(map[string]interface{}{
			"from": attrs.From,
			"to":   attrs.To,
		}).Debug(p.Body)
		threemaID := strings.ToUpper(strings.Split(attrs.To, "@")[0])
		err := account.Send(threemaID, p.Body)
		if err != nil {
			msg := xmpp.NewMessage("chat", "", attrs.From, "", "en")
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
