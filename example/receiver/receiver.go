package receiver

type User struct {
	Name string
}

func (user *User) InitReceiver(
	getHeader func(string) string,
	getEnv func(string) string,
) (err error) {

	user.Name = getEnv("USER_NAME")
	if name := getHeader("UserName"); name != "" {
		user.Name = name
	}

	return nil
}

func (user User) Greeting() (greeting string, err error) {
	return "Hello, my name is " + user.Name, nil
}
