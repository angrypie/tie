package tasks

import (
	"errors"

	"github.com/angrypie/tie/upgrade"
)

func Binary(pkg, dest string, services []Service) error {
	client, err := upgrade.Client(pkg)
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.Name == pkg {
			continue
		}
		ok := client.Replace(service.Name, service.Name+"/tie_client")
		if !ok {
			return errors.New("Imports not replaced")
		}
	}
	client.Write()
	if err != nil {
		return err
	}

	server, err := upgrade.Server(pkg + "/tie_bin")
	if err != nil {
		return err
	}
	if len(server.Functions) > 0 {
		server.Write()
		if err != nil {
			return err
		}
	}

	return nil
}
