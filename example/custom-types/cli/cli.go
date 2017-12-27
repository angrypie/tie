package cli

import "github.com/angrypie/tie/example/custom-types/register"

func Register() {
	user := register.User{"Paul", "Super secret"}
	register.NewUser(user)
}
