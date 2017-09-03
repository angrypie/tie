package tasks

import "io/ioutil"

func ReadConfigFile(dest string) error {
	buf, err := ioutil.ReadFile(dest)
	if err != nil {
		return err
	}

	return Config(buf)
}
