package useradmin

type UserConfig interface {
	UserName() string
}

type User struct {
	Name string
}

func NewUser(config UserConfig) (model *User, err error) {
	model.Name = config.UserName()
	return
}

type AdminConfig interface {
	AdminRole() string
}

type Admin struct {
	Role string
}

func NewAdmin(config AdminConfig) (model *Admin, err error) {
	model.Role = config.AdminRole()
	return
}
