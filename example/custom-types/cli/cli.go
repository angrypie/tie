package main

import (
	"log"

	"github.com/angrypie/tie/example/custom-types/register"
)

func main() {
	user := register.User{Name: "Paul", Password: "PaulPassword"}
	err := register.NewUser(user)
	if err != nil {
		log.Println("ERR new user", err)
	}
}
