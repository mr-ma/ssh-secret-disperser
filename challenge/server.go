package challenge

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/mr-ma/ssh-secret-disperser/secret"
	"github.com/mr-ma/ssh-secret-disperser/user"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/ssh"
)

const SERVER_TIMEOUT time.Duration = 2 * time.Minute

//Server challenge-based secret disclosure
type Server struct {
	Config                 *ssh.ServerConfig
	Store                  *secret.Storage
	quit                   chan interface{}
	RequestedSecretIndexes map[string]int
	users                  *user.ActiveDirectory
}

//NewChallengeServer constructor of server
func NewChallengeServer() *Server {
	s := new(Server)
	s.quit = make(chan interface{})
	s.Config = s.loadConfig()
	s.Store = secret.MakeStorage()
	s.users = user.MakeUserDirectory()
	//plug in the connection handler
	s.users.ConnectListeners = make([]user.ConnectionListenerFunctionType, 1)
	s.users.ConnectListeners[0] = func(username string) {
		//check if we have two users connected and if they qualify for a secret disclosure
		s.CheckForPairMatches()
	}
	s.RequestedSecretIndexes = make(map[string]int)
	return s
}

func (s *Server) ResetState() {
	s.RequestedSecretIndexes = make(map[string]int)
}

//Start the challenge server
func (s *Server) Start() error {
	log.Println("Starting server...")
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	for {
		select {
		case <-s.quit:
			return nil
		default:
			nConn, err := listener.Accept()
			if err != nil {
				log.Fatal("failed to accept incoming connection: ", err)
			}
			log.Printf("New SSH connection from %s", nConn.RemoteAddr())

			go s.serveSSHConnection(nConn)

		}
	}
}

func (s *Server) serveSSHConnection(nConn net.Conn) {

	//set timeouts on the connection
	nConn.SetDeadline(time.Now().Add(SERVER_TIMEOUT))

	sshcon, chans, reqs, err := ssh.NewServerConn(nConn, s.Config)
	if err != nil {
		// log.Fatal("failed to handshake: ", err)
		log.Println("client authentication failed...")
		return
	}
	isValidUser, user := s.users.GetSysUser(sshcon.User())
	if !isValidUser {
		panic("authenticated but user does not exist!")
	}

	// Discard all global out-of-band Requests
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			log.Println(newChannel.ChannelType())
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, _, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				req.Reply(req.Type == "shell", nil)
			}
		}(reqs)
		s.users.ConnectUser(user, s.RequestedSecretIndexes[sshcon.User()], channel)
	}
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

//Stop server
func (s *Server) Stop() error {
	log.Println("Stopping server...")
	if !IsClosed(s.quit) {
		close(s.quit)
	}
	s.users.Stop()
	return nil
}

func (s *Server) challenge(conn ssh.ConnMetadata,
	client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	answers, err := client(conn.User(), "",
		[]string{"Please choose a number between 1 to 10:",
			"Please enter OTP code:"},
		[]bool{true, true})

	if err != nil {
		return nil, err
	}

	if len(answers) == 2 && answers[0] != "" && answers[1] != "" {

		//check if the user exists in the sys users
		isValidUser, sysUser := s.users.GetSysUser(conn.User())
		if !isValidUser {
			log.Println("Error. User does not exist")
			return nil, errors.New("Error. User does not exist")
		}
		//check if the user is already connected
		isAlreadyConnected := s.users.IsUserAlreadyConnected(conn.User())
		if isAlreadyConnected {
			log.Println("Error. User is already connected")
			return nil, errors.New("Error. User is already connected")
		}
		requestedSecretIndex, err := strconv.Atoi(answers[0])
		if err != nil {
			log.Println("Error. Cannot convert the requested secret index to int")
			return nil, err
		}
		//if OTP code is good connect user
		valid := totp.Validate(answers[1], sysUser.OTPSecret)
		if !valid {
			log.Println("Error. OTP authentication failed")
			return nil, errors.New("Error. OTP authentication failed")
		}

		s.RequestedSecretIndexes[conn.User()] = requestedSecretIndex
		return nil, nil
	}
	return nil, fmt.Errorf("Failed to answer questions")
}

func (s *Server) loadConfig() *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		KeyboardInteractiveCallback: s.challenge,
		MaxAuthTries:                1,
	}
	private, err := ssh.ParsePrivateKey(secret.PrivateKey)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)
	return config
}
