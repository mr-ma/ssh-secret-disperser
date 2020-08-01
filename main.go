package main

import (
	"os"
	"os/signal"

	"github.com/mr-ma/ssh-secret-disperser/challenge"
)

func main() {

	//graceful shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	s := challenge.NewChallengeServer()
	go s.Start()

	select {
	case <-c:
		s.Stop()
	}
}
