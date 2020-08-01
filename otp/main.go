package main

import (
	"fmt"

	"github.com/pquerna/otp/totp"
)

func main() {

	key1, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "Example.com",
		AccountName: "test1@example.com",
	})
	key2, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "Example.com",
		AccountName: "test2@example.com",
	})

	fmt.Printf("test1 secret:%s\n", key1.Secret())
	fmt.Printf("test2 secret:%s\n", key2.Secret())
}
