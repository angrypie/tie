package helth

import "log"

func Check(name string) error {
	log.Println(name, "is ok")
	return nil
}
