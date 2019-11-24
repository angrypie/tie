package recconf

import "github.com/angrypie/tie/example/receiver-config/useradmin"

type APIConfig interface {
	UserName() string
	AdminRole() string
}

type API struct {
	user  *useradmin.User
	admin *useradmin.Admin
}

func NewAPI(user *useradmin.User, admin *useradmin.Admin, config APIConfig) (api *API, err error) {
	if user != nil {
		user, err = useradmin.NewUser(config)
		if err != nil {
			return
		}
	}

	if admin != nil {
		admin, err = useradmin.NewAdmin(config)
		if err != nil {
			return
		}
	}
	return &API{user, admin}, nil
}

type UserModel struct {
	Name string
	user *useradmin.User
}

type UserModelConfig interface {
	Name() string
}

func NewUserModel(api *API, config UserModelConfig) (model *UserModel, err error) {
	name := config.Name()
	return &UserModel{
		Name: name,
		user: api.user,
	}, nil
}
