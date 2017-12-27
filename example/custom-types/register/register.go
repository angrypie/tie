package register

import "log"

type User struct {
	Name     string
	Password string
}

func NewUser(user User) (ok bool, err error) {
	log.Println("New user registered", user.Name)
	return true, nil
}
