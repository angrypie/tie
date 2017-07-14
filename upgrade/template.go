package upgrade

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func (upgrade *ServerUpgrade) initServerUpgrade(p *parser.Parser) error {
	header, err := template.MakeServerHeader(p)
	if err != nil {
		return err
	}
	upgrade.Server.Write(header)
	if err != nil {
		return err
	}
	return nil
}

func (upgrade *ServerUpgrade) addServerMain(p *parser.Parser) error {
	main, err := template.MakeServerMain(p.Package)
	if err != nil {
		return err
	}
	upgrade.Server.Write(main)
	if err != nil {
		return err
	}
	return nil
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
