package types

type Service struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Type  string `yaml:"type"`
	Port  string `yaml:"port"`
}

type ConfigFile struct {
	Services []Service `yaml:"services"`
	Path     string    `yaml:"path"`
}
