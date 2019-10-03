package component

import (
	"strings"

	"github.com/bdlm/log"
	"gosrc.io/xmpp/stanza"
)

func (c *Config) sender(packets chan stanza.Packet) {
	log.Debugf("start xmpp sender for: %s", c.Host)
	for packet := range packets {
		if p := c.sending(packet); p != nil {
			c.xmpp.Send(p)
		}
	}
}

func (c *Config) fixAddr(addr string) string {
	if addr == "" {
		return c.Host
	}
	if strings.Contains(addr, "{{DOMAIN}}") {
		return strings.Replace(addr, "{{DOMAIN}}", c.Host, 1)
	}
	if !strings.Contains(addr, "@") {
		return addr + "@" + c.Host
	}
	return addr
}

func (c *Config) sending(packet stanza.Packet) stanza.Packet {
	logger := log.WithField("type", c.Type)
	switch p := packet.(type) {
	case stanza.Presence:
		p.From = c.fixAddr(p.From)
		if p.To != "" {
			p.To = c.fixAddr(p.To)
		}
		return p
	case stanza.Message:
		p.From = c.fixAddr(p.From)
		if p.To != "" {
			p.To = c.fixAddr(p.To)
		}
		if c.XMPPDebug {
			logger.WithFields(map[string]interface{}{
				"from": p.From,
				"to":   p.To,
				"id":   p.Id,
			}).Debug(p.XMPPFormat())
		}
		return p
	default:
		log.Warn("ignoring packet:", packet)
		return nil
	}
}
