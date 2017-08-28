package upgrade

import (
	"fmt"
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

func (upgrade *ClientUpgrade) Clean() error {
	//upgrade.Package.Path
	fs := afero.NewOsFs()
	folder := upgrade.Parser.Package.Path + "/tie_bin"
	err := fs.RemoveAll(folder)
	if err != nil {

		return err
	}
	return nil
}

func (upgrade *ServerUpgrade) Clean() error {
	fs := afero.NewOsFs()
	path := upgrade.Package.Path

	err := fs.RemoveAll(path + "/tie_server")
	if err != nil {
		return err
	}

	err = fs.RemoveAll(path + "/tie_client")
	if err != nil {
		return err
	}
	return nil
}

func (upgrader *Upgrader) Write() error {
	fs := afero.NewOsFs()
	path := upgrader.Parser.Package.Path

	//#1
	err := fs.MkdirAll(path+"/tie_server", 0755)
	if err != nil {
		return err
	}
	err = afero.WriteFile(
		fs,
		path+"/tie_server/server.go",
		upgrader.Server.Bytes(),
		0644,
	)
	if err != nil {
		return err
	}
	//#2

	err = fs.MkdirAll(path+"/tie_upgraded", 0755)
	if err != nil {
		return err
	}

	files := upgrader.Parser.ToFiles()
	for index, file := range files {
		err = afero.WriteFile(
			fs,
			fmt.Sprintf("%s/tie_upgraded/%d.go", path, index),
			file.Bytes(),
			0644,
		)
		if err != nil {
			return err
		}
	}

	//#3
	err = fs.MkdirAll(path+"/tie_client", 0755)
	if err != nil {
		return err
	}
	err = afero.WriteFile(
		fs,
		path+"/tie_client/client.go",
		upgrader.Client.Bytes(),
		0644,
	)
	if err != nil {
		return err
	}
	//TODO write tie_upgraded

	return nil
}
