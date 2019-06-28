package component

import (
	"encoding/xml"

	"github.com/bdlm/log"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func (c *Config) handleDiscoInfo(s xmpp.Sender, p stanza.Packet) {
	iq, ok := p.(stanza.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	discoInfo, ok := iq.Payload.(*stanza.DiscoInfo)
	if !ok {
		return
	}
	attrs := iq.Attrs
	iq = stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, To: attrs.From, From: attrs.To, Id: attrs.Id})

	payload := stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: stanza.NSDiscoInfo,
			Local: "query",
		},
		Features: []stanza.Feature{
			{Var: stanza.NSDiscoInfo},
			{Var: stanza.NSDiscoItems},
			{Var: stanza.NSMsgReceipts},
			{Var: stanza.NSMsgChatMarkers},
			{Var: stanza.NSMsgChatStateNotifications},
		},
	}
	if discoInfo.Node == "" {
		payload.Identity = append(payload.Identity, stanza.Identity{
			Name:     c.Type,
			Category: "gateway",
			Type:     "service",
		})
	}
	iq.Payload = &payload
	log.WithFields(map[string]interface{}{
		"type": c.Type,
		"from": s,
		"to":   attrs.To,
	}).Debug("disco info")
	s.Send(iq)
}

func (c *Config) handleDiscoItems(s xmpp.Sender, p stanza.Packet) {
	iq, ok := p.(stanza.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	discoItems, ok := iq.Payload.(*stanza.DiscoItems)
	if !ok {
		return
	}
	attrs := iq.Attrs
	iq = stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, To: attrs.From, From: attrs.To, Id: attrs.Id})

	payload := stanza.DiscoItems{}
	if discoItems.Node == "" {
		payload.Items = []stanza.DiscoItem{
			{Name: c.Type, JID: c.Host, Node: "node1"},
		}
	}
	iq.Payload = &payload

	log.WithFields(map[string]interface{}{
		"type": c.Type,
		"from": s,
		"to":   attrs.To,
	}).Debug("disco items")
	s.Send(iq)
}
func (c *Config) handleIQ(s xmpp.Sender, p stanza.Packet) {
	iq, ok := p.(stanza.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	xError := stanza.Err{
		Code:   501,
		Reason: "feature-not-implemented",
		Type:   "cancel",
	}
	resp := iq.MakeError(xError)
	attrs := iq.Attrs

	log.WithFields(map[string]interface{}{
		"type": c.Type,
		"from": s,
		"to":   attrs.To,
	}).Debugf("ignore: %s", iq.Payload)
	s.Send(resp)
}
func (c *Config) handleMessage(s xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		return
	}
	if c.XMPPDebug {
		log.WithFields(map[string]interface{}{
			"type": c.Type,
			"from": s,
			"to":   msg.To,
			"id":   msg.Id,
		}).Debug(msg.XMPPFormat())
	}
	c.comp.Send(p)
}
