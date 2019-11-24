package receiver

import (
	"fmt"
)

type greeter struct {
	phrase string
}

func (g *greeter) greet(name string) string {
	return fmt.Sprintf("%s %s", g.phrase, name)
}

type Provider struct {
	greeter *greeter
}

func (p *Provider) Hello(name string) (str string, err error) {
	return p.greeter.greet(name), nil
}

func NewProvider(getEnv func(string) string) (p *Provider, err error) {
	phrase := "Hello brah"
	if p := getEnv("DEFAULT_PHRASE"); p != "" {
		phrase = p
	}
	return &Provider{&greeter{phrase}}, nil
}

type User struct {
	Name    string
	greeter *greeter
}

func NewUser(identity Identity, p *Provider, getHeader func(string) string) (user *User, err error) {
	return &User{
		Name:    getHeader("UserName"),
		greeter: p.greeter,
	}, nil
}

func (user *User) Hello() (greeting string, err error) {
	return user.greeter.greet(user.Name), nil
}

func Hello(name string) (str string, err error) {
	g := greeter{}
	return g.greet(name), nil
}

type Identity struct {
	Name string
}
