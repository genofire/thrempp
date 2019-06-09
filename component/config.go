package component

import (
	"gosrc.io/xmpp"
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

func (c *Config) Start() (err error) {
	out, err := c.comp.Connect()
	if err != nil {
		return
	}
	c.xmpp, err = xmpp.NewComponent(xmpp.ComponentOptions{
		Domain:   c.Host,
		Secret:   c.Secret,
		Address:  c.Connection,
		Name:     c.Type,
		Category: "gateway",
		Type:     "service",
	})
	if err != nil {
		return
	}
	cm := xmpp.NewStreamManager(c.xmpp, nil)
	err = cm.Start()
	if err != nil {
		return
	}

	go c.sender(out)
	go c.receiver()

	return nil
}
