package user

import "golang.org/x/crypto/ssh"

type Session struct {
	ActiveUser           User
	Channel              ssh.Channel
	RequestedSecretIndex int
}

func (session *Session) CloseSession() error {
	err := session.Channel.Close()
	return err
}
