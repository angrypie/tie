package receiver

import (
	"fmt"
)

func Hello(name string) (str string, err error) {
	g := greeter{}
	return g.greet(name), nil
}

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

func NewUser(p *Provider, name string, getHeader func(string) string) (user *User, err error) {
	g := p.greeter
	if phrase := getHeader("Hello-Phrase"); phrase != "" {
		g = &greeter{phrase}
	}
	return &User{
		Name:    name,
		greeter: g,
	}, nil
}

func (user *User) Hello(text string) (greeting string, err error) {
	return user.greeter.greet(user.Name + ": " + text), nil
}

type Guest struct{}

func (guest *Guest) Hello() (greeting string, err error) {
	return "Hello guest", nil
}
