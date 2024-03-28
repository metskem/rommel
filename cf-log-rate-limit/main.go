package main

import (
	"context"
	"github.com/cloudfoundry-community/go-cfclient/v3/client"
	"github.com/cloudfoundry-community/go-cfclient/v3/config"
	"github.com/cloudfoundry-community/go-cfclient/v3/resource"
	"log"
	"os"
	"time"
)

var (
	apiAddress, clientId, clientSecret string
	oldLimit                           = 25000
	newLimit                           = 10000
)

func initConfig() {
	configComplete := true
	if apiAddress = os.Getenv("API_ADDRESS"); apiAddress == "" {
		log.Println("missing envvar: API_ADDRESS")
		configComplete = false
	}
	if clientId = os.Getenv("CLIENT_ID"); clientId == "" {
		log.Println("missing envvar: CLIENT_ID")
		configComplete = false
	}
	if clientSecret = os.Getenv("CLIENT_SECRET"); clientSecret == "" {
		log.Println("missing envvar: CLIENT_SECRET")
		configComplete = false
	}
	if !configComplete {
		os.Exit(1)
	}
}

func main() {
	initConfig()
	ctx := context.TODO()
	if cfConfig, err := config.NewUserPassword(apiAddress, clientId, clientSecret); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		if cfClient, err := client.New(cfConfig); err != nil {
			log.Fatalf("failed to create new client: %s", err)
		} else {
			// cf client is there, now do the real work
			// first get all orgs and spaces
			log.Println("getting all spaces...")
			if allSpaces, err := cfClient.Spaces.ListAll(ctx, nil); err != nil {
				log.Fatalf("failed to get all spaces: %s", err)
			} else {
				log.Println("getting all orgs...")
				if allOrgs, err := cfClient.Organizations.ListAll(ctx, nil); err != nil {
					log.Fatalf("failed to get all orgs: %s", err)
				} else {
					// put them in a map for easy lookup by guid
					spacesMap := make(map[string]*resource.Space)
					for _, space := range allSpaces {
						spacesMap[space.GUID] = space
					}
					orgsMap := make(map[string]*resource.Organization)
					for _, org := range allOrgs {
						orgsMap[org.GUID] = org
					}
					log.Println("getting all processes...")
					if allProcesses, err := cfClient.Processes.ListAll(ctx, nil); err != nil {
						log.Fatalf("failed to Processes.ListAll: %s", err)
					} else {
						log.Printf("found %d processes\n\n", len(allProcesses))
						var counter int
						for ix, process := range allProcesses {
							if process.LogRateLimitInBytesPerSecond == oldLimit {
								if app, err := cfClient.Applications.Get(ctx, process.Relationships.App.Data.GUID); err != nil {
									log.Printf("failed to get app with guid %s: %s", process.Relationships.App.Data.GUID, err)
								} else {
									currentSpace := spacesMap[app.Relationships.Space.Data.GUID]
									currentOrg := orgsMap[currentSpace.Relationships.Organization.Data.GUID]
									if currentOrg.Name != "system" && currentOrg.Name != "cf-services" {
										log.Printf("%d (%d) - processguid:%s %s %s - %s/%s/%s\n", ix, counter, process.GUID, process.Type, app.State, currentOrg.Name, currentSpace.Name, app.Name)
										counter++
										if _, err := cfClient.Processes.Scale(ctx, process.GUID, &resource.ProcessScale{LogRateLimitInBytesPerSecond: &newLimit}); err != nil {
											log.Printf("failed to scale: %s", err)
										}
									}
								}
							}
							time.Sleep(200 * time.Millisecond)
						}
					}
				}
			}
		}
	}
}
