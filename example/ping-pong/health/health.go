package health

import "log"

func Check(name string) (err error) {
	log.Println(name, "is ok")
	return nil
}
