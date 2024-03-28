package main

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/hazelcast/hazelcast-go-client/types"
	"log"
	"os"
	"time"
)

var dryRun = true

func main() {
	if len(os.Args) < 5 || len(os.Args) > 6 {
		log.Println("I need 4 arguments: <cluster address> (10.2.3.4:5701), <clustername> (cluster-xyz), <username>, <password>, dryRun (optional, default true)")
		os.Exit(1)
	} else {
		config := hazelcast.Config{}
		config.ClientName = "panzer-hzutil"
		config.Cluster.Network.SetAddresses(os.Args[1])
		config.Cluster.Name = os.Args[2]
		config.Cluster.Security.Credentials.Username = os.Args[3]
		config.Cluster.Security.Credentials.Password = os.Args[4]
		if len(os.Args) > 5 && os.Args[5] == "false" {
			dryRun = false
		}
		config.Logger.Level = logger.WarnLevel
		log.SetOutput(os.Stdout)
		ctx := context.TODO()

		var err error
		var client *hazelcast.Client
		if client, err = hazelcast.StartNewClientWithConfig(ctx, config); err != nil {
			log.Printf("failed to start new client: %s\n", err)
		} else {
			var distObjects []types.DistributedObjectInfo
			var findCount, destroyCount int
			if distObjects, err = client.GetDistributedObjectsInfo(ctx); err != nil {
				log.Printf("failed to get distributed objects: %s\n", err)
			} else {
				for _, distObject := range distObjects {
					var mappie *hazelcast.Map
					mappie, err = client.GetMap(ctx, distObject.Name)
					if err != nil {
						log.Printf("failed to get map: %s\n", err)
					} else {
						if mapSize, err := mappie.Size(ctx); mapSize == 0 {
							findCount++
							_, _ = fmt.Fprintf(os.Stderr, "%d - %s\n", findCount, mappie.Name())
							if !dryRun {
								if err = mappie.Destroy(ctx); err != nil {
									log.Printf("      failed to destroy : %s\n", err)
								} else {
									destroyCount++
								}
							}
						}
					}
					time.Sleep(time.Millisecond * 50)
				}
				log.Printf("found %d distributed objects @ %s (%d destroyed)\n", len(distObjects), config.Cluster.Name, destroyCount)
			}
		}
		if err = client.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown client: %s\n", err)
		}
	}
}
