package api

import (
	"errors"
	"fmt"
)

var humans map[string]Human

func InitService() (err error) {
	humans = make(map[string]Human)
	fmt.Println("Server started")
	return
}

type Human struct {
	Name   string
	Age    int
	Gender string
}

type CreateHumanResponse struct {
	Msg string `json:"msg"`
	Human
}

func CreateHuman(name, gender string, age int) (response CreateHumanResponse, err error) {
	_, ok := humans[name]
	if ok {
		return CreateHumanResponse{}, errors.New("already exist")
	}

	humans[name] = Human{Name: name, Age: age, Gender: gender}

	return CreateHumanResponse{
		Msg:   fmt.Sprintf("Human %s created.", name),
		Human: humans[name],
	}, nil
}

func GetHuman(name string) (human Human, err error) {
	human, ok := humans[name]
	if !ok {
		err = errors.New("not found")
	}
	return
}

func DeleteHuman(name string) (err error) {
	_, ok := humans[name]
	if !ok {
		err = errors.New("not found")
		return
	}
	delete(humans, name)
	return
}
