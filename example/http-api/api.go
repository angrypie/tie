package api

import "fmt"

type CreateHumanDTO struct {
	Name     string
	Age      int
	Gender   string
	Location Location
}

type Location struct {
	City string
}

func CreateHuman(requestDTO CreateHumanDTO) (msg string, err error) {
	name, location := requestDTO.Name, requestDTO.Location.City
	return fmt.Sprintf("Human %s from %s created.", name, location), nil
}

func DeleteHuman(name string) (msg string, err error) {
	return fmt.Sprintf("Human %s deleted", name), nil
}
