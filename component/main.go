package component

import (
	"github.com/bdlm/log"
	"gosrc.io/xmpp"
)

type Component interface {
	Connect() (chan xmpp.Packet, error)
	Send(xmpp.Packet)
}

// Connect function with config to get DB connection interface
type Connect func(config map[string]interface{}) (Component, error)

var components = map[string]Connect{}

func AddComponent(name string, c Connect) {
	components[name] = c
}

func Load(configs []Config) {
	for _, config := range configs {
		f, ok := components[config.Type]
		if !ok {
			log.Warnf("it was not possible to find a component with type '%s'", config.Type)
			continue
		}
		comp, err := f(config.Special)
		if err != nil {
			log.WithField("type", config.Type).Panic(err)
		}
		config.comp = comp
		log.WithField("type", config.Type).Infof("component for %s started", config.Host)
	}
}
