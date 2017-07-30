package upgrade

import (
	"strconv"

	"github.com/spf13/afero"
)

func (upgrade *ServerUpgrade) Write() error {
	fs := afero.NewOsFs()
	path := upgrade.Package.Path
	err := fs.MkdirAll(path+"/tie_server", 0755)
	if err != nil {
		return err
	}
	err = afero.WriteFile(
		fs,
		path+"/tie_server/server.go",
		upgrade.Server.Bytes(),
		0644,
	)
	if err != nil {
		return err
	}

	err = fs.MkdirAll(path+"/tie_client", 0755)
	if err != nil {
		return err
	}
	err = afero.WriteFile(
		fs,
		path+"/tie_client/client.go",
		upgrade.Client.Bytes(),
		0644,
	)
	if err != nil {
		return err
	}

	return nil
}
func (upgrade *ClientUpgrade) Write() error {
	//upgrade.Package.Path
	fs := afero.NewOsFs()
	folder := upgrade.Parser.Package.Path + "/tie_bin"
	err := fs.MkdirAll(folder, 0755)
	if err != nil {
		return err
	}

	for index, file := range upgrade.Client {
		err := afero.WriteFile(
			fs,
			folder+"/"+strconv.Itoa(index)+".go",
			file.Bytes(),
			0644,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
