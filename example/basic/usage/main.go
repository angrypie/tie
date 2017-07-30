package main

import (
	"log"

	"github.com/angrypie/tie/example/basic"
)

func main() {
	res, err := basic.Mul(5, 9)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Result: ", res)
	}
}
