package receiver

type User struct {
	Name string
}

func (user *User) InitReceiver(getHeader func(string) string) (err error) {
	user.Name = getHeader("UserName")
	return nil
}

func (user User) Greeting() (greeting string, err error) {
	return "Hello, my name is " + user.Name, nil
}
