package tasks

import (
	"fmt"
	"regexp"

	"github.com/spf13/afero"
)

func CleanBinary(dest string) error {
	fs := afero.NewRegexpFs(afero.NewOsFs(), regexp.MustCompile("*.run"))
	err := fs.RemoveAll(fmt.Sprintf("%s/.*", dest))
	if err != nil {
		return err
	}
	return nil
}
