package threema

import (
	"dev.sum7.eu/genofire/thrempp/component"
	"gosrc.io/xmpp"
)

type Threema struct {
	component.Component
}

func NewThreema(config map[string]interface{}) (component.Component, error) {
	return &Threema{}, nil
}

func (t *Threema) Connect() (chan xmpp.Packet, error) {
	c := make(chan xmpp.Packet)
	return c, nil
}

func init() {
	component.AddComponent("threema", NewThreema)
}
