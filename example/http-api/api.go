package api

import (
	"errors"
	"fmt"
)

func InitService() (err error) {
	fmt.Println("Server started")
	return
}

type CreateHumanRequest struct {
	Name   string
	Age    int
	Gender string
}

type CreateHumanResponse struct {
	Msg string `json:"msg"`
}

func CreateHuman(id string, request CreateHumanRequest) (response CreateHumanResponse, err error) {
	req := &request
	name, id := req.Name, id
	if name == "paul" {
		return CreateHumanResponse{}, errors.New("already exist")
	}
	return CreateHumanResponse{
		Msg: fmt.Sprintf("Human %s (%s) created.", name, id),
	}, nil
}

func DeleteHuman(name string, gender string, age int) (msg string, err error) {
	return fmt.Sprintf("Human %s (%s) deleted", name, gender), nil
}
