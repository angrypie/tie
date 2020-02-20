package template

type Module interface {
	Name() string
	Generate() (Package, error)
	Deps() []Module
}

type Package struct {
	Name string
	Path string
	Code string
}

func TraverseModules(module Module, cb func(p Module) error) (err error) {
	err = cb(module)
	if err != nil {
		return err
	}

	for _, dep := range module.Deps() {
		err = TraverseModules(dep, cb)
		if err != nil {
			return err
		}
	}

	return
}
