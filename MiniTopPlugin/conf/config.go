package conf

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
)

var (
	ApiAddr string
	ShardId = "MiniTopPlugin"
)

func EnvironmentComplete(cliConnection plugin.CliConnection) bool {
	envComplete := true
	var err error
	if ApiAddr, err = cliConnection.ApiEndpoint(); err != nil {
		envComplete = false
		fmt.Printf("Error getting API endpoint: %v\n", err)
	}
	return envComplete
}
