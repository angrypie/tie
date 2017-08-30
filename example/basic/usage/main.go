package main

import (
	"fmt"

	"github.com/angrypie/tie/example/basic"
)

func main() {
	res, err := basic.Mul(7, 6)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%d+%d=%d\n", 7, 5, res)
	}
}
