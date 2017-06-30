package upgrade

import (
	"fmt"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func initServerUpgrade(upgrade *ServerUpgarde) {
	header := template.ServerHeader
	upgrade.Server = append([]byte(header))
}

func addApiEndpoint(file []byte, function parser.Function) {
	apiWrapper := fmt.Sprintf(template.ApiWrapper, function.Name)
	file = append([]byte(apiWrapper))
}

func addApiClient(file []byte, function parser.Function) {

}
