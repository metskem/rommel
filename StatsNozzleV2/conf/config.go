package conf

import (
	"fmt"
	"os"
)

var (
	ApiAddr = os.Getenv("API_ADDR")
	ShardId = os.Getenv("SHARD_ID")
	Client  = os.Getenv("CLIENT_ID")
	Secret  = os.Getenv("CLIENT_SECRET")
)

func EnvironmentComplete() bool {
	envComplete := true
	if ApiAddr == "" {
		fmt.Println("missing envvar: API_ADDR")
		envComplete = false
	}
	if ShardId == "" {
		fmt.Println("missing envvar: SHARD_ID")
		envComplete = false
	}
	if Client == "" {
		fmt.Println("missing envvar: CLIENT_ID")
		envComplete = false
	}
	if Secret == "" {
		fmt.Println("missing envvar: CLIENT_SECRET")
		envComplete = false
	}
	return envComplete
}
