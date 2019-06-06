package component

import (
	"github.com/bdlm/log"
	"gosrc.io/xmpp"
)

func (c *Config) sender(packets chan xmpp.Packet) {
	for packet := range packets {
		if p := c.sending(packet); p != nil {
			c.xmpp.Send(p)
		}
	}
}

func (c *Config) sending(packet xmpp.Packet) xmpp.Packet {
	logger := log.WithField("type", c.Type)
	switch p := packet.(type) {
	case xmpp.Message:
		if p.PacketAttrs.From == "" {
			p.PacketAttrs.From = c.Host
		} else {
			p.PacketAttrs.From += "@" + c.Host
		}
		if c.XMPPLog {
			logger.WithFields(map[string]interface{}{
				"from": p.PacketAttrs.From,
				"to":   p.PacketAttrs.To,
				"id":   p.PacketAttrs.Id,
			}).Debug(p.XMPPFormat())
		}
		return p
	default:
		log.Warn("ignoring packet:", packet)
		return nil
	}
}
