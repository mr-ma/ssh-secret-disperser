package challenge

import (
	"log"

	"golang.org/x/crypto/ssh/terminal"
)

func (s *Server) CheckForPairMatches() {
	countUsers := len(s.users.OnlineUsers)
	if countUsers < 2 {
		//we need to wait for the other user
		return
	}
	if countUsers > 2 {
		panic("We accept two users max")
	}
	canReveal := true

	//get requested indexes out of the map
	var requestedIndexes []int
	for _, v := range s.users.OnlineUsers {
		requestedIndexes = append(requestedIndexes, v.RequestedSecretIndex)
	}

	firstRequestedIndex := requestedIndexes[0]
	for i := 0; i < countUsers; i++ {
		if firstRequestedIndex != requestedIndexes[i] {
			canReveal = false
			break
		}
	}
	if canReveal {
		s.DiscloseSecret(firstRequestedIndex)
	} else {
		log.Println("Cannot disclose the secret because indexes do not match")
	}

	//wheather we disclosed a secret or not, disconnect all
	s.ResetState() //clear state
	s.users.DisconnectAll()
}

func (s *Server) DiscloseSecret(index int) error {
	log.Printf("Disclosing Secret index:%d\n", index)
	secret, err := s.Store.GetSecret(index)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, session := range s.users.OnlineUsers {
		terminal := terminal.NewTerminal(session.Channel, "> ")
		_, err := terminal.Write([]byte(secret))
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}
	return nil
}
