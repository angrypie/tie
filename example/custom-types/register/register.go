package register

import "fmt"

type User struct {
	Name, Password string
}

func NewUser(user User) (err error) {
	fmt.Printf("New user %s with password %s\n", user.Name, user.Password)
	return nil
}
