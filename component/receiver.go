package component

import (
	"github.com/bdlm/log"
	"gosrc.io/xmpp"
)

func (c *Config) receiver() {
	for {
		packet, err := c.xmpp.ReadPacket()
		if err != nil {
			log.WithField("type", c.Type).Panicf("connection closed%s", err)
			return
		}
		p, back := c.receive(packet)
		if p == nil {
			continue
		}
		if back {
			c.xmpp.Send(p)
		} else {
			c.comp.Send(p)
		}
	}
}

func (c *Config) receive(packet xmpp.Packet) (xmpp.Packet, bool) {
	logger := log.WithField("type", c.Type)

	switch p := packet.(type) {
	case xmpp.IQ:
		attrs := p.PacketAttrs
		loggerIQ := logger.WithFields(map[string]interface{}{
			"from": attrs.From,
			"to":   attrs.To,
		})

		switch inner := p.Payload[0].(type) {
		case *xmpp.DiscoInfo:
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
						{Var: xmpp.NSSpaceXEP0184Receipt},
						{Var: xmpp.NSSpaceXEP0333ChatMarkers},
					},
				}
				iq.AddPayload(&payload)
				loggerIQ.Debug("disco info")
				return iq, true
			}
		case *xmpp.DiscoItems:
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
				loggerIQ.Debug("disco items")
				return iq, true
			}
		default:
			logger.Debug("ignoring iq packet", inner)
			xError := xmpp.Err{
				Code:   501,
				Reason: "feature-not-implemented",
				Type:   "cancel",
			}
			reply := p.MakeError(xError)

			return reply, true
		}

	case xmpp.Message:
		logger.WithFields(map[string]interface{}{
			"from": p.PacketAttrs.From,
			"to":   p.PacketAttrs.To,
			"id":   p.PacketAttrs.Id,
		}).Debug(p.XMPPFormat())

		return packet, false

	case xmpp.Presence:
		logger.Debug("received presence:", p.Type)

	default:
		logger.Debug("ignoring packet:", packet)
	}
	return nil, false
}
