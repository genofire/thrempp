package component

import (
	"github.com/bdlm/log"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

type Config struct {
	Type       string
	Host       string
	Connection string
	Secret     string
	XMPPDebug  bool `toml:"xmpp_debug"`
	Special    map[string]interface{}

	xmpp *xmpp.Component
	comp Component
}

func (c *Config) Start() error {
	out, err := c.comp.Connect()
	if err != nil {
		return err
	}

	router := xmpp.NewRouter()
	router.NewRoute().IQNamespaces(stanza.NSDiscoInfo).HandlerFunc(c.handleDiscoInfo)
	router.NewRoute().IQNamespaces(stanza.NSDiscoItems).HandlerFunc(c.handleDiscoItems)
	router.HandleFunc("iq", c.handleIQ)
	router.HandleFunc("message", c.handleMessage)

	c.xmpp, err = xmpp.NewComponent(xmpp.ComponentOptions{
		Domain:   c.Host,
		Secret:   c.Secret,
		Address:  c.Connection,
		Name:     c.Type,
		Category: "gateway",
		Type:     "service",
	}, router)
	if err != nil {
		return err
	}
	cm := xmpp.NewStreamManager(c.xmpp, nil)
	go func() {
		if err := cm.Run(); err != nil {
			log.WithField("host", c.Host).Panicf("xmpp stream not started: %s", err)
		}
	}()
	go c.sender(out)

	return nil
}
