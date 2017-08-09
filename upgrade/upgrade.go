package upgrade

import (
	"bytes"

	"github.com/angrypie/tie/parser"
)

type ServerUpgrade struct {
	Server    bytes.Buffer
	Client    bytes.Buffer
	Package   *parser.Package
	Functions []*parser.Function
}

type ClientUpgrade struct {
	Client []bytes.Buffer
	Parser *parser.Parser
}

//Server scan package for public function declarations and
//generates RPC API wrappers for this functions and RPC client for this API
func Server(pkg string) (upgrade *ServerUpgrade, err error) {
	upgrade = &ServerUpgrade{}
	p := parser.NewParser()
	err = p.Parse(pkg)
	if err != nil {
		return upgrade, err
	}
	functions, err := p.GetFunctions()
	upgrade.Functions = functions
	if err != nil {
		return upgrade, err
	}
	upgrade.initServerUpgrade(p)

	for _, function := range functions {
		err = upgrade.addApiEndpoint(function)
		if err != nil {
			return upgrade, err
		}
		err = upgrade.addApiClient(function)
		if err != nil {
			return upgrade, err
		}
	}

	err = upgrade.addServerMain(p)
	if err != nil {
		return upgrade, err
	}

	upgrade.Package = p.Package
	return upgrade, err
}

//Client scan package for using methad calls that are API endpoints in another packages
//and replace this calls with API calls
func Client(pkg string) (upgrade *ClientUpgrade, err error) {
	upgrade = &ClientUpgrade{}
	p := parser.NewParser()
	err = p.Parse(pkg)
	if err != nil {
		return upgrade, err
	}
	upgrade.Parser = p
	return upgrade, err
}

func (upgrade *ClientUpgrade) Replace(from, to string) (ok bool) {
	ok, upgrade.Client = upgrade.Parser.ReplaceImport(from, to)
	return ok
}
