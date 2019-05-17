package api

import (
	"errors"
	"fmt"
)

type CreateHumanDTO struct {
	Name   string
	Age    int
	Gender string
}

type HumanCreatedDTO struct {
	Msg string
	Err error
}

func CreateHuman(requestDto CreateHumanDTO) (msg string, err error) {
	req := &requestDto
	name, age := req.Name, req.Age
	if name == "paul" {
		return "", errors.New("already exist")
	}
	//name, location := requestDTO.Name, requestDTO.Location.City
	return fmt.Sprintf("Human %s (%d) created.", name, age), nil
}

func DeleteHuman(name string) (msg string, err error) {
	return fmt.Sprintf("Human %s deleted", name), nil
}
