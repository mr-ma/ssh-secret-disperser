package user

import (
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

const SESSION_TIMEOUT time.Duration = 2 * time.Minute

type ConnectionListenerFunctionType = func(string)

type ActiveDirectory struct {
	OnlineUsers         map[string]Session
	sysUsers            map[string]User
	store               Storage
	handlequit          chan interface{}
	timeoutquit         chan interface{}
	disconnecttimeout   chan string
	DisconnectListeners []ConnectionListenerFunctionType
	ConnectListeners    []ConnectionListenerFunctionType
}

func MakeUserDirectory() *ActiveDirectory {
	ud := new(ActiveDirectory)
	ud.OnlineUsers = make(map[string]Session)
	ud.sysUsers = make(map[string]User)
	ud.handlequit = make(chan interface{})
	ud.timeoutquit = make(chan interface{})
	ud.disconnecttimeout = make(chan string)
	//TODO: plug in a user storage
	ud.store = DummyStore{}
	ud.sysUsers, _ = ud.store.LoadUsers()
	go ud.handleTimeoutDisconnect()
	return ud
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
func (ud *ActiveDirectory) Stop() {

	//exit gracefully
	if !IsClosed(ud.handlequit) {
		close(ud.handlequit)
	}
	if !IsClosed(ud.timeoutquit) {
		close(ud.timeoutquit)
	}
	log.Println("gracefully stopped activedirectory")
}

func (ud *ActiveDirectory) GetSysUser(name string) (exists bool, user User) {
	user, found := ud.sysUsers[name]
	return found, user
}

func (ud *ActiveDirectory) IsUserAlreadyConnected(name string) bool {
	_, found := ud.OnlineUsers[name]
	return found
}

func (ud *ActiveDirectory) informDisconnectToListeners(username string) {
	for _, listener := range ud.DisconnectListeners {
		listener(username)
	}
}
func (ud *ActiveDirectory) informConnectToListeners(username string) {
	for _, listener := range ud.ConnectListeners {
		listener(username)
	}
}

func (ud *ActiveDirectory) handleTimeoutDisconnect() {
	go func() {
		for {
			select {
			case <-ud.handlequit:
				log.Println("Quit handleTimeoutDisconnect listener")
				return
			case username := <-ud.disconnecttimeout:
				if ud.IsUserAlreadyConnected(username) {
					log.Printf("Disconnecting:%s\n", username)
					ud.DisconnectUser(username)
					ud.informDisconnectToListeners(username)
				} else {
					log.Printf("Disconnect Timeout reached... for an already disconnected user %s\n", username)
				}
				break
			}
		}
	}()
}

func (ud *ActiveDirectory) DisconnectAll() {
	for key := range ud.OnlineUsers {
		ud.DisconnectUser(key)
	}
}

func (ud *ActiveDirectory) DisconnectUser(username string) bool {
	if ud.IsUserAlreadyConnected(username) {
		session, err := ud.GetUserSession(username)
		if err != nil {
			panic(err)
		}
		session.CloseSession()
		delete(ud.OnlineUsers, username)
		return true
	}
	return false
}
func (ud *ActiveDirectory) ConnectUser(user User, requestedSecretIndex int, channel ssh.Channel) error {
	log.Printf("UserDirectory ConnectUser %s\n", user.Name)
	ud.OnlineUsers[user.Name] = Session{
		ActiveUser:           user,
		Channel:              channel,
		RequestedSecretIndex: requestedSecretIndex,
	}
	//start timeout
	ud.watchTimeout(user.Name)
	ud.informConnectToListeners(user.Name)
	return nil
}

func (ud *ActiveDirectory) watchTimeout(username string) {
	log.Printf("UserDirectory watchTimeout %s\n", username)
	ticker := time.NewTicker(SESSION_TIMEOUT)
	go func() {
		for {
			select {
			case <-ud.timeoutquit:
				log.Printf("watchTimeout: Quit timeout listener user:%s", username)
				return
			case <-ticker.C:
				log.Printf("watchTimeout: Timeout reached... initiating disconnect user:%s", username)
				ud.disconnecttimeout <- username
				return
			}
		}
	}()
}
func (ud *ActiveDirectory) GetUserSession(name string) (Session, error) {
	user, found := ud.OnlineUsers[name]
	if !found {
		return Session{}, errors.New("Error. Did not find user")
	} else {
		return user, nil
	}

}
