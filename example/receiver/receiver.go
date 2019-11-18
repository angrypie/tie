package receiver

import (
	"fmt"
)

type KV = func(string) string
type greeter struct{}

func (g *greeter) greet(name string) string {
	return fmt.Sprintf("Hello brah %s\n", name)
}

type Provider struct {
	greeter *greeter
}

func (p *Provider) InitReceiver() (err error) {
	p.greeter = &greeter{}
	return
}

func (p *Provider) User(getHeader KV) (user *User, err error) {
	name := getHeader("UserName")
	return &User{name, p.greeter}, nil
}

type User struct {
	Name    string
	greeter *greeter
}

func (user *User) Hello() (greeting string, err error) {
	return user.greeter.greet(user.Name), nil
}
