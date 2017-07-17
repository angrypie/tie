package upgrade

import (
	"bytes"
	"log"

	"github.com/angrypie/tie/parser"
)

type Upgrade struct {
}

type ServerUpgrade struct {
	Server  bytes.Buffer
	Client  bytes.Buffer
	Package *parser.Package
}

//Server scan package for public function declarations and
//generates RPC API wrappers for this functions, and RPC client for this API
func Server(pkg string) (upgrade ServerUpgrade, err error) {
	p := parser.NewParser()
	err = p.Parse(pkg)
	if err != nil {
		return upgrade, err
	}
	functions, err := p.GetFunctions()
	if err != nil {
		return upgrade, err
	}
	upgrade.initServerUpgrade(p)
	for _, function := range functions {
		log.Println(function.Name)
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

	//add client header
	//add client rpc calls

	upgrade.Package = p.Package
	return upgrade, err
}

//Client scan package for using methad calls that are API endpoints in another packages
//and replace this calls with API calls
func Client(pckg string) (upgrade ServerUpgrade, err error) {
	return upgrade, err
}
