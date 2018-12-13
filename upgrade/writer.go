package upgrade

import (
	"fmt"

	"github.com/spf13/afero"
)

//writeHelper creates directory for package and write files.
func writeHelper(path, dir string, files ...[]byte) error {
	fs := afero.NewOsFs()
	fullPath := fmt.Sprintf("%s/%s", path, dir)

	err := fs.MkdirAll(fullPath, 0755)
	if err != nil {
		return err
	}

	for index, file := range files {
		err = afero.WriteFile(
			fs,
			fmt.Sprintf("%s/%d.go", fullPath, index),
			file,
			0644,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

//Write writes packages from buffer to filesystem
func (upgrader *Upgrader) Write() error {
	path := upgrader.Parser.Package.Path

	err := writeHelper(path, "tie_server", upgrader.Server.Bytes())
	if err != nil {
		return err
	}

	err = writeHelper(path, "tie_upgraded", upgrader.Parser.ToFiles()...)
	if err != nil {
		return err
	}

	err = writeHelper(path, "tie_client", upgrader.Client.Bytes())
	if err != nil {
		return err
	}

	return nil
}

//Clean removes files and directories created by Write method
func (upgrader *Upgrader) Clean() error {
	fs := afero.NewOsFs()
	path := upgrader.Parser.Package.Path
	tmpDirs := []string{"tie_server", "tie_client", "tie_upgraded"}

	for _, dir := range tmpDirs {
		if err := fs.RemoveAll(fmt.Sprintf("%s/%s", path, dir)); err != nil {
			return err
		}
	}

	return nil
}
