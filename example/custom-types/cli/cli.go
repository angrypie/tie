package main

import "github.com/angrypie/tie/example/custom-types/register"

func main() {
	user := register.User{"Paul", "PaulPassword"}
	register.NewUser(user)
}
