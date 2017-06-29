package upgrade

type Upgrade struct {
}

type ServerUpgarde struct {
	Server []byte
	Client []byte
}

//Server scan package for public function declarations and
//generates RPC API wrappers for this functions, and RPC client for this API
func Server(pkg string) (upgrade ServerUpgarde, err error) {
	functions, _ := parse.FindFunctions(pkg)
	initServerUpgrade(upgrade)
	for _, function := range functions {
		addApiEndpoint(upgrade.Server, function)
		addApiClient(upgrade.Client, function)
	}
	return upgrade, err
}

//Client scan package for using methad calls that are API endpoints in another packages
//and replace this calls with API calls
func Client(pckg string) (upgrade ServerUpgarde, err error) {
	return upgrade, err
}
