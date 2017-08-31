package tasks

import (
	"github.com/angrypie/tie/upgrade"
)

func Replace(pkg string, services []Service) (*upgrade.Upgrader, error) {
	upgrader := upgrade.NewUpgrader(pkg)

	err := upgrader.Parse()
	if err != nil {
		return nil, err
	}

	imports := make([]string, len(services))

	for i, service := range services {
		imports[i] = service.Name
	}

	err = upgrader.Replace(imports)
	if err != nil {
		return nil, err
	}

	err = upgrader.Make()
	if err != nil {
		return nil, err
	}

	err = upgrader.Write()
	if err != nil {
		return nil, err
	}

	return upgrader, nil
}
