package threema

import "errors"

type ThreemaAccount struct {
	ID string
}

func (t *Threema) getAccount(jid string) *ThreemaAccount {
	return &ThreemaAccount{}
}

func (a *ThreemaAccount) Send(to string, msg string) error {
	if a.ID == "" {
		return errors.New("It was not possible to send, becaouse we have no account for you.\nPlease generate one, by sending `generate` to gateway")
	}
	return nil
}
