package tasks

import (
	"log"
	"os/exec"

	"github.com/spf13/afero"
)

//TODO Move files to tmp
func CleanBinary(dest string) (removed []string, err error) {
	fs := afero.NewOsFs()
	files, err := afero.Glob(fs, "*.run")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		isDir, err := afero.IsDir(fs, file)
		if err != nil {
			return nil, err
		}
		if isDir {
			continue
		}
		err = fs.Remove(file)
		if err != nil {
			return nil, err
		}
		removed = append(removed, file)
	}
	return removed, nil
}

func InstallGoDependencies() error {
	deps := []string{
		"github.com/labstack/echo",
		"github.com/dgrijalva/jwt-go",
		"github.com/smallnest/rpcx/...",
	}
	for _, dep := range deps {
		_, err := exec.Command("go", "get", dep).Output()
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
