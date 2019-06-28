package threema

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gosrc.io/xmpp/stanza"
)

func (a *Account) FileToXMPP(from string, msgID uint64, ext string, data []byte) (stanza.Message, error) {
	msgIDStr := strconv.FormatUint(msgID, 10)
	msg := stanza.Message{
		Attrs: stanza.Attrs{
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
	msg.Extensions = append(msg.Extensions, stanza.OOB{URL: url})
	return msg, nil
}
