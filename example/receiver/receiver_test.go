package receiver

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReceiver(t *testing.T) {
	user := User{}
	greeting, err := user.Hello()
	require.NoError(t, err)
	log.Println(greeting)

}
