package upgrade

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

func (upgrade *Upgrader) initServerUpgrade(p *parser.Parser) error {
	serverHeader, err := template.MakeServerHeader(p)
	if err != nil {
		return err
	}

	clientHeader, err := template.MakeClientHeader(p)
	if err != nil {
		return err
	}

	upgrade.Server.Write(serverHeader)
	if err != nil {
		return err
	}
	upgrade.Client.Write(clientHeader)
	if err != nil {
		return err
	}

	return nil
}

func (upgrade *Upgrader) addServerMain(p *parser.Parser) error {
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

func (upgrade *Upgrader) addApiEndpoint(function *parser.Function) error {
	wrapper, err := template.MakeApiWrapper(function)
	if err != nil {
		return err
	}

	client, err := template.MakeApiClient(function)
	if err != nil {
		return err
	}

	_, err = upgrade.Server.Write(wrapper)
	if err != nil {
		return err
	}

	_, err = upgrade.Client.Write(client)
	if err != nil {
		return err
	}
	return nil
}
