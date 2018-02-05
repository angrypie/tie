package tasks

import (
	"github.com/angrypie/tie/upgrade"
)

//upgradeWithServices crate new upgrader for pkg and upgrade with services
func upgradeWithServices(current Service, services []Service) (*upgrade.Upgrader, error) {
	pkg := current.Name
	upgrader := upgrade.NewUpgrader(pkg, current.Type)

	imports := make([]string, len(services))
	for i, service := range services {
		imports[i] = service.Name
	}

	err := upgrader.Upgrade(imports)
	if err != nil {
		return nil, err
	}

	return upgrader, nil
}
