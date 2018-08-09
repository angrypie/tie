package tasks

import (
	"github.com/spf13/afero"
)

func CleanBinary(dest string) error {
	fs := afero.NewOsFs()
	files, err := afero.Glob(fs, "*.run")
	if err != nil {
		return err
	}

	for _, file := range files {
		err := fs.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}
