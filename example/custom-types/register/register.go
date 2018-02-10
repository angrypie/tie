package register

import "fmt"

type User struct {
	Name, Password string
}

func NewUser(user User) (ok bool, err error) {
	fmt.Printf("New user %s with password %s", user.Name, user.Password)
	return true, nil
}
