package upgrade

import "io/ioutil"

func (upgrade *ServerUpgrade) Write() error {
	//upgrade.Package.Path
	err := ioutil.WriteFile(
		upgrade.Package.Path+"/tie_server/server.go",
		upgrade.Server.Bytes(), 0644,
	)
	if err != nil {
		return err
	}

	//upgrade.Client
	//write to path/tie_client/client.go
	//upgrade.Server
	return nil
}
