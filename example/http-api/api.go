package api

import (
	"errors"
	"fmt"
)

type CreateHumanRequest struct {
	Name   string
	Age    int
	Gender string
}

type CreateHumanResponse struct {
	Msg string `json:"msg"`
	Err error  `json:"err"`
}

func CreateHuman(id string, requestDto CreateHumanRequest) (responseDTO CreateHumanResponse) {
	req := &requestDto
	name, id := req.Name, id
	if name == "paul" {
		return CreateHumanResponse{Err: errors.New("already exist")}
	}
	return CreateHumanResponse{
		Msg: fmt.Sprintf("Human %s (%s) created.", name, id),
	}
}

func DeleteHuman(name string, gender string, age int) (msg string, err error) {
	return fmt.Sprintf("Human %s (%s) deleted", name, gender), nil
}
