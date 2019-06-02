package threema

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gosrc.io/xmpp"
)

func (a *Account) FileToXMPP(from string, msgID uint64, ext string, data []byte) (xmpp.Message, error) {
	msgIDStr := strconv.FormatUint(msgID, 10)
	msg := xmpp.Message{
		PacketAttrs: xmpp.PacketAttrs{
			Id:   msgIDStr,
			From: from,
			To:   a.XMPP.String(),
		},
	}
	url := fmt.Sprintf("%s/%d.%s", a.threema.httpUploadURL, msgID, ext)
	path := fmt.Sprintf("%s/%d.%s", a.threema.httpUploadPath, msgID, ext)
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		msg.Body = "unable to save file on transport to forward"
		return msg, err
	}
	msg.Body = url
	msg.X = &xmpp.MsgXOOB{URL: url}
	return msg, nil
}
