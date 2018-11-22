package main

import (
	"fmt"

	"github.com/angrypie/tie/example/basic/sum"
)

func main() {
	res, err := sum.Sum(7, 6)

	//Even though Mul method not returning error, Tie internals may return it.
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%d+%d=%d\n", 7, 5, res)
	}
}
