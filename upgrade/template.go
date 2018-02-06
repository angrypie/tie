package upgrade

import (
	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/template"
)

//initServerUpgrade writes header templates to RPC client and server buffers
func (upgrader *Upgrader) initServerUpgrade(p *parser.Parser) error {
	serverHeader, err := template.MakeServerHeader(p)
	if err != nil {
		return err
	}

	clientHeader, err := template.MakeClientHeader(p)
	if err != nil {
		return err
	}

	clientTypes, err := template.MakeClientTypes(p)
	if err != nil {
		return err
	}

	upgrader.Server.Write(serverHeader)
	if err != nil {
		return err
	}

	upgrader.Client.Write(clientHeader)
	if err != nil {
		return err
	}

	upgrader.Client.Write(clientTypes)
	if err != nil {
		return err
	}

	return nil
}

//addServerMain writes main function template to RPC server package
func (upgrader *Upgrader) addServerMain(p *parser.Parser, functions []*parser.Function) error {
	main, err := template.MakeServerMain(p, functions)
	if err != nil {
		return err
	}

	upgrader.Server.Write(main)
	if err != nil {
		return err
	}

	return nil
}

//addApiEndpoint  adds API handler to RPC, or HTTP server and client method to client package
func (upgrader *Upgrader) addApiEndpoint(function *parser.Function) error {
	wrapper, err := template.MakeApiWrapper(function)
	if err != nil {
		return err
	}

	client, err := template.MakeApiClient(function)
	if err != nil {
		return err
	}

	_, err = upgrader.Server.Write(wrapper)
	if err != nil {
		return err
	}

	_, err = upgrader.Client.Write(client)
	if err != nil {
		return err
	}
	return nil
}
