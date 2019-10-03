package threema

import (
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/o3ma/o3"
	"gosrc.io/xmpp"
)

func strFromThreemaGroup(header *o3.GroupMessageHeader) string {
	cid := strings.ToLower(header.CreatorID.String())
	gid := strings.ToLower(base32.StdEncoding.EncodeToString(header.GroupID[:]))
	return fmt.Sprintf("%s-%s", cid, gid)
}
func jidFromThreemaGroup(sender string, header *o3.GroupMessageHeader) string {
	return fmt.Sprintf("%s@{{DOMAIN}}/%s", strFromThreemaGroup(header), sender)
}
func jidToThreemaGroup(jidS string) (string, *o3.GroupMessageHeader) {
	jid, err := xmpp.NewJid(jidS)
	if err != nil {
		return "", nil
	}
	node := strings.ToUpper(jid.Node)
	a := strings.SplitN(node, "-", 2)
	if len(a) != 2 {
		return "", nil
	}
	header := &o3.GroupMessageHeader{
		CreatorID: o3.NewIDString(a[0]),
	}

	result, err := base32.StdEncoding.DecodeString(a[1])
	if err != nil {
		return "", nil
	}
	copy(header.GroupID[:], []byte(result))

	return jid.Resource, header
}
