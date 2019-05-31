package component

import (
	"gosrc.io/xmpp"
)

type Config struct {
	Type       string
	Host       string
	Connection string
	Secret     string
	Special    map[string]interface{}

	xmpp *xmpp.Component
	comp Component
}

func (c *Config) Start() error {
	c.xmpp = &xmpp.Component{Host: c.Host, Secret: c.Secret}
	err := c.xmpp.Connect(c.Connection)
	if err != nil {
		return err
	}
	out, err := c.comp.Connect()
	if err != nil {
		return err
	}

	go c.recieve(out)
	go c.sender()

	return nil
}

func (c *Config) recieve(chan xmpp.Packet) {
}
func (c *Config) sender() {
}
