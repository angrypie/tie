package template

import (
	"fmt"
	"testing"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
)

func TestMain(t *testing.T) {
	info := PackageInfo{
		IsStopService: true,
		IsInitService: true,
		Service: &types.Service{
			Type: "http",
			Auth: "supersecret",
		},
		Functions: []*parser.Function{
			&parser.Function{
				Name: "DelUser",
				Arguments: []parser.Field{
					parser.Field{Name: "name", Type: "string", Package: ""},
					parser.Field{Name: "force", Type: "bool", Package: ""},
				},
				Results: []parser.Field{
					parser.Field{Name: "id", Type: "string"},
					parser.Field{Name: "err", Type: "error"},
				},
			},
			&parser.Function{
				Name: "GetUsers",
				Arguments: []parser.Field{
					parser.Field{Name: "requestDTO", Type: "GetUserRequest", Package: "daemon"},
				},
				Results: []parser.Field{
					parser.Field{Name: "id", Type: "string"},
					parser.Field{Name: "force", Type: "bool"},
					parser.Field{Name: "err", Type: "error"},
				},
			},
		},
	}
	serverMain, _ := GetServerMain(&info)
	fmt.Println(serverMain)
}
