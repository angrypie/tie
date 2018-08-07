package types

type Service struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Type  string `yaml:"type"`
	Port  string `yaml:"port"`
	Auth  string `yaml:"auth"`
}

type ConfigFile struct {
	Services []Service `yaml:"services"`
	Path     string    `yaml:"path"`
}
