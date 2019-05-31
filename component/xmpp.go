package component

import (
	"github.com/bdlm/log"
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
	logger := log.WithField("type", c.Type)
	for {
		logger.Debug("wait fo recieve")
		packet, err := c.xmpp.ReadPacket()
		if err != nil {
			logger.Panicf("connection closed%s", err)
			return
		}
		logger.Debug("recieve")

		switch p := packet.(type) {
		case xmpp.IQ:
			attrs := p.PacketAttrs

			switch inner := p.Payload[0].(type) {
			case *xmpp.DiscoInfo:
				logger.Debug("Disco Info")
				if p.Type == "get" {
					iq := xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")
					var identity xmpp.Identity
					if inner.Node == "" {
						identity = xmpp.Identity{
							Name:     c.Type,
							Category: "gateway",
							Type:     "service",
						}
					}

					payload := xmpp.DiscoInfo{
						Identity: identity,
						Features: []xmpp.Feature{
							{Var: "http://jabber.org/protocol/disco#info"},
							{Var: "http://jabber.org/protocol/disco#item"},
						},
					}
					iq.AddPayload(&payload)

					_ = c.xmpp.Send(iq)
				}
			case *xmpp.DiscoItems:
				logger.Debug("DiscoItems")
				if p.Type == "get" {
					iq := xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")

					var payload xmpp.DiscoItems
					if inner.Node == "" {
						payload = xmpp.DiscoItems{
							Items: []xmpp.DiscoItem{
								{Name: c.Type, JID: c.Host, Node: "node1"},
							},
						}
					}
					iq.AddPayload(&payload)
					_ = c.xmpp.Send(iq)
				}
			default:
				logger.Warn("ignoring iq packet", inner)
				xError := xmpp.Err{
					Code:   501,
					Reason: "feature-not-implemented",
					Type:   "cancel",
				}
				reply := p.MakeError(xError)
				_ = c.xmpp.Send(&reply)
			}

		case xmpp.Message:
			logger.Info("Received message:", p.Body)

		case xmpp.Presence:
			logger.Info("Received presence:", p.Type)

		default:
			logger.Warn("ignoring packet:", packet)
		}
	}
}
