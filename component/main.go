package component

import (
	"github.com/bdlm/log"
	"gosrc.io/xmpp/stanza"
)

type Component interface {
	Connect() (chan stanza.Packet, error)
	Send(stanza.Packet)
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
		err = config.Start()
		if err != nil {
			log.WithField("type", config.Type).Panic(err)
		}
		log.WithField("type", config.Type).Infof("component for %s started", config.Host)
	}
}
