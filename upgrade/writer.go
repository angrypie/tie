package upgrade

import (
	"io/ioutil"
	"strconv"
)

func (upgrade *ServerUpgrade) Write() error {
	//upgrade.Package.Path
	err := ioutil.WriteFile(
		upgrade.Package.Path+"/tie_server/server.go",
		upgrade.Server.Bytes(), 0644,
	)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		upgrade.Package.Path+"/tie_client/client.go",
		upgrade.Client.Bytes(), 0644,
	)
	if err != nil {
		return err
	}

	//upgrade.Client
	//write to path/tie_client/client.go
	//upgrade.Server
	return nil
}
func (upgrade *ClientUpgrade) Write() error {
	//upgrade.Package.Path
	for index, file := range upgrade.Client {
		err := ioutil.WriteFile(
			upgrade.Parser.Package.Path+"/tie_bin/"+strconv.Itoa(index)+".go",
			file.Bytes(), 0644,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
