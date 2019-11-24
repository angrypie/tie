package health

import "log"

func Check(name string) error {
	log.Println(name, "is ok")
	return nil
}
