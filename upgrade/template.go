package upgrade

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func (upgrade *ServerUpgrade) initServerUpgrade() {
	header := template.ServerHeader
	upgrade.Server.Write([]byte(header))
}

func (upgrade *ServerUpgrade) addApiEndpoint(function *parser.Function) error {
	wrapper, err := template.MakeApiWrapper(function)
	if err != nil {
		return err
	}
	_, err = upgrade.Server.Write(wrapper)
	if err != nil {
		return err
	}
	return nil
}

func (upgrade *ServerUpgrade) addApiClient(function *parser.Function) error {
	return nil
}
