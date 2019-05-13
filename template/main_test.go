package template

import (
	"fmt"
	"testing"

	"github.com/angrypie/tie/parser"
	"github.com/angrypie/tie/types"
)

func TestMain(t *testing.T) {
	info := PackageInfo{
		Path:          "github.com/someservice/daemon",
		IsStopService: true,
		IsInitService: true,
		Service: &types.Service{
			Type: "http",
			Auth: "supersecret",
		},
		Functions: []*parser.Function{
			&parser.Function{
				Name: "GetUsers",
				Arguments: []parser.Field{
					parser.Field{Name: "", Type: "GetUserRequest", Package: "emon"},
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
