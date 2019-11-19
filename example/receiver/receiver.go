package receiver

import (
	"fmt"
	"log"
)

type KV = func(string) string
type greeter struct {
	phrase string
}

func (g *greeter) greet(name string) string {
	return fmt.Sprintf("%s %s/n", g.phrase, name)
}

type Provider struct {
	greeter *greeter
}

func (p *Provider) InitReceiver(getEnv func(string) string) (err error) {
	phrase := "Hello brah"
	log.Println("lol")
	if p := getEnv("DEFAULT_PHRASE"); p != "" {
		phrase = p
	}
	p.greeter = &greeter{phrase}
	return
}

func (p *Provider) User(getHeader string) (user *User, err error) {
	name := getHeader
	return &User{name, p.greeter}, nil
}

type User struct {
	Name    string
	greeter *greeter
}

func (user *User) Hello() (greeting string, err error) {
	return user.greeter.greet(user.Name), nil
}

func Hello(name string) (str string, err error) {
	g := greeter{}
	return g.greet(name), nil
}
