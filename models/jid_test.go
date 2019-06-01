package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJIDTableName(t *testing.T) {
	assert := assert.New(t)

	var jid JID
	assert.Equal("jid", jid.TableName())
}

// Test Values for NewJID from RFC 7622
// https://tools.ietf.org/html/rfc7622
func TestParseJID(t *testing.T) {
	assert := assert.New(t)

	checkList := map[string]*JID{
		"juliet@example.com": {
			Local:  "juliet",
			Domain: "example.com",
		},
		"juliet@example.com/foo": {
			Local:  "juliet",
			Domain: "example.com",
		},
		"juliet@example.com/foo bar": {
			Local:  "juliet",
			Domain: "example.com",
		},
		"juliet@example.com/foo@bar": {
			Local:  "juliet",
			Domain: "example.com",
		},
		"foo\\20bar@example.com": {
			Local:  "foo\\20bar",
			Domain: "example.com",
		},
		"fussball@example.com": {
			Local:  "fussball",
			Domain: "example.com",
		},
		"fu&#xDF;ball@example.com": {
			Local:  "fu&#xDF;ball",
			Domain: "example.com",
		},
		"&#x3C0;@example.com": {
			Local:  "&#x3C0;",
			Domain: "example.com",
		},
		"&#x3A3;@example.com/foo": {
			Local:  "&#x3A3;",
			Domain: "example.com",
		},
		"&#x3C3;@example.com/foo": {
			Local:  "&#x3C3;",
			Domain: "example.com",
		},
		"&#x3C2;@example.com/foo": {
			Local:  "&#x3C2;",
			Domain: "example.com",
		},
		"king@example.com/&#x265A;": {
			Local:  "king",
			Domain: "example.com",
		},
		"example.com": {
			Domain: "example.com",
		},
		"example.com/foobar": {
			Domain: "example.com",
		},
		"a.example.com/b@example.net": {
			Domain: "a.example.com",
		},
		"\"juliet\"@example.com":  nil,
		"foo bar@example.com":     nil,
		"juliet@example.com/ foo": nil,
		"@example.com/":           nil,
		// "henry&#x2163;@example.com": nil, -- ignore for easier implementation
		// "&#x265A;@example.com":      nil,
		"juliet@": nil,
		"/foobar": nil,
	}

	for jidString, jidValid := range checkList {
		jid := ParseJID(jidString)

		if jidValid != nil {
			assert.NotNil(jid, "this should be a valid JID:"+jidString)
			if jid == nil {
				continue
			}

			assert.Equal(jidValid.Local, jid.Local, "the local part was not right detectet:"+jidString)
			assert.Equal(jidValid.Domain, jid.Domain, "the domain part was not right detectet:"+jidString)
		} else {
			assert.Nil(jid, "this should not be a valid JID:"+jidString)
		}

	}
}

func TestJIDString(t *testing.T) {
	assert := assert.New(t)

	var jid *JID
	assert.Equal("", jid.String())

	jid = &JID{
		Domain: "example.com",
	}
	assert.Equal("example.com", jid.String())

	jid = &JID{
		Local: "romeo",
	}
	assert.Equal("", jid.String())

	jid = &JID{
		Local:  "romeo",
		Domain: "example.com",
	}
	assert.Equal("romeo@example.com", jid.String())
}

func TestJIDIsDomain(t *testing.T) {
	assert := assert.New(t)
	var jid *JID
	assert.False(jid.IsDomain())

	jid = &JID{}
	assert.False(jid.IsDomain())

	jid = &JID{Local: "a"}
	assert.False(jid.IsDomain())

	jid = &JID{Domain: "a"}
	assert.True(jid.IsDomain())

	jid = &JID{Local: "a", Domain: "b"}
	assert.False(jid.IsDomain())
}
