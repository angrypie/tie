package pets

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
)

// pets database-ish
var pets = make(map[string]struct{})

// use JSON logging when run by Up (including `up start`).
func InitService() (err error) {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(json.Default)
	}
	return nil
}

// curl http://localhost:3000/
func Get() (resp string, err error) {
	log.Info("list pets")

	if len(pets) == 0 {
		return "no pets", nil
	}

	return fmt.Sprintf("%v", pets), nil
}

// curl -d Tobi http://localhost:3000/
// curl -d Loki http://localhost:3000/
// curl -d Jane http://localhost:3000/
func Post(name string) (resp string, err error) {
	pets[name] = struct{}{}
	log.WithField("name", name).Info("add pet")

	return fmt.Sprintf("welcome to the family %s!\n", name), nil
}

// curl -X DELETE http://localhost:3000/Tobi
// curl -X DELETE http://localhost:3000/Loki
// curl -X DELETE http://localhost:3000/Jane
func Delete(name string) (resp string, err error) {
	log.WithField("name", name).Info("remove pet")
	delete(pets, name)
	return fmt.Sprintf("removed %s!\n", name), nil

}
