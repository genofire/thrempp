package component

import (
	"encoding/xml"

	"github.com/bdlm/log"
	"gosrc.io/xmpp"
)

func (c *Config) handleDiscoInfo(s xmpp.Sender, p xmpp.Packet) {
	iq, ok := p.(xmpp.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	discoInfo, ok := iq.Payload.(*xmpp.DiscoInfo)
	if !ok {
		return
	}
	attrs := iq.PacketAttrs
	iq = xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")

	payload := xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		},
		Features: []xmpp.Feature{
			{Var: xmpp.NSDiscoInfo},
			{Var: xmpp.NSDiscoItems},
			{Var: xmpp.NSMsgReceipts},
			{Var: xmpp.NSMsgChatMarkers},
			{Var: xmpp.NSMsgChatStateNotifications},
		},
	}
	if discoInfo.Node == "" {
		payload.Identity = xmpp.Identity{
			Name:     c.Type,
			Category: "gateway",
			Type:     "service",
		}
	}
	iq.Payload = &payload
	log.WithFields(map[string]interface{}{
		"type": c.Type,
		"from": s,
		"to":   attrs.To,
	}).Debug("disco info")
	s.Send(iq)
}

func (c *Config) handleDiscoItems(s xmpp.Sender, p xmpp.Packet) {
	iq, ok := p.(xmpp.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	discoItems, ok := iq.Payload.(*xmpp.DiscoItems)
	if !ok {
		return
	}
	attrs := iq.PacketAttrs
	iq = xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")

	payload := xmpp.DiscoItems{}
	if discoItems.Node == "" {
		payload.Items = []xmpp.DiscoItem{
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
func (c *Config) handleIQ(s xmpp.Sender, p xmpp.Packet) {
	iq, ok := p.(xmpp.IQ)
	if !ok || iq.Type != "get" {
		return
	}
	xError := xmpp.Err{
		Code:   501,
		Reason: "feature-not-implemented",
		Type:   "cancel",
	}
	resp := iq.MakeError(xError)
	attrs := iq.PacketAttrs

	log.WithFields(map[string]interface{}{
		"type": c.Type,
		"from": s,
		"to":   attrs.To,
	}).Debugf("ignore: %s", iq.Payload)
	s.Send(resp)
}
func (c *Config) handleMessage(s xmpp.Sender, p xmpp.Packet) {
	msg, ok := p.(xmpp.Message)
	if !ok {
		return
	}
	if c.XMPPDebug {
		log.WithFields(map[string]interface{}{
			"type": c.Type,
			"from": s,
			"to":   msg.PacketAttrs.To,
			"id":   msg.PacketAttrs.Id,
		}).Debug(msg.XMPPFormat())
	}
	c.comp.Send(p)
}
