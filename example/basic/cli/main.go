package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/angrypie/tie/example/basic/sum"
)

func main() {
	a, b := argByNumber(1), argByNumber(2)
	res, err := sum.Sum(a, b)

	//Even though Mul method not returning error, Tie internals may return it.
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%d+%d=%d\n", a, b, res)
	}
}

func argByNumber(i int) int {
	if len(os.Args) <= i {
		return 0
	}
	a, _ := strconv.Atoi(os.Args[i])
	return a
}
