package main

import "github.com/angrypie/tie/example/custom-types/register"

func main() {
	user := register.User{Name: "Paul", Password: "PaulPassword"}
	register.NewUser(user)
}
