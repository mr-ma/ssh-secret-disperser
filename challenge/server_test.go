package challenge

import (
	"log"
	"testing"
	"time"

	"github.com/mr-ma/ssh-secret-disperser/secret"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/ssh"
)

func ClientConfigHelper(username, optsecret, secretindex string) *ssh.ClientConfig {
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(secret.PublicKey)
	if err != nil {
		log.Panic(err)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(sshuser, instruction string,
				questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				otpcode, _ := totp.GenerateCode(optsecret, time.Now())
				log.Println(otpcode)
				answers[0] = secretindex
				answers[1] = otpcode
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.FixedHostKey(publicKey),
	}
	return config
}

func TestConnection(t *testing.T) {
	config := ClientConfigHelper("user1", "L2NJFMNEJRBI2SNZQ2HUJNGRDCZEGTGM", "1")
	client, err := ssh.Dial("tcp", "0.0.0.0:2022", config)
	if err != nil {
		t.Fail()
	}
	session, err := client.NewSession()
	if err != nil {
		t.Fail()
	}
	defer session.Close()

	teardown()
}

func TestFailWrongOTPSecret(t *testing.T) {
	//provide false otp secret
	config := ClientConfigHelper("user1", "FALSE OTP SECRET", "1")
	_, err := ssh.Dial("tcp", "0.0.0.0:2022", config)
	//dial must fail
	if err == nil {
		t.Fail()
	}
	teardown()
}

func TestSecretDisclosure(t *testing.T) {
	config1 := ClientConfigHelper("user1", "L2NJFMNEJRBI2SNZQ2HUJNGRDCZEGTGM", "1")
	config2 := ClientConfigHelper("user2", "LQDE6HPJHG55LXAHQ4LNEN2J2G6UIFHC", "1")
	client1, err := ssh.Dial("tcp", "0.0.0.0:2022", config1)
	if err != nil {
		t.Fail()
	}
	client2, err := ssh.Dial("tcp", "0.0.0.0:2022", config2)
	if err != nil {
		t.Fail()
	}
	session1, err := client1.NewSession()
	if err != nil {
		t.Fail()
	}
	session2, err := client2.NewSession()
	if err != nil {
		t.Fail()
	}
	defer session1.Close()
	defer session2.Close()

	teardown()
}

var s *Server

func init() {
	s = NewChallengeServer()
	go s.Start()

}
func teardown() {
	s.Stop()
}
