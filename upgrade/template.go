package upgrade

import (
	"fmt"
	"go/types"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func (upgrade *ServerUpgrade) initServerUpgrade() {
	header := template.ServerHeader
	upgrade.Server.Write([]byte(header))
}

func (upgrade *ServerUpgrade) addApiEndpoint(function *parser.Function) {
	var args string
	for _, arg := range function.Arguments {
		args = arg.Name + " " + types.ExprString(arg.Type) + ", "
	}
	apiWrapper := fmt.Sprintf(template.ApiWrapper, function.Name, args)
	upgrade.Server.Write([]byte(apiWrapper))
}

func (upgrade *ServerUpgrade) addApiClient(function *parser.Function) {

}
