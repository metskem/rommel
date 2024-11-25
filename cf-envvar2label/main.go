package main

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"log"
	"os"
)

var (
	apiAddress, clientId, clientSecret string
	labels2Set                         = []string{"SPLUNK_INDEX", "RABO_CI", "F2S_DISABLE_LOGGING"}
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <cf_home_dir>", os.Args[0])
	}
	ctx := context.TODO()
	if cfConfig, err := config.NewFromCFHomeDir(os.Args[1]); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		if cfClient, err := client.New(cfConfig); err != nil {
			log.Fatalf("failed to create new client: %s", err)
		} else {
			// cf client is there, now do the real work
			log.Println("getting all apps...")
			if allApps, err := cfClient.Applications.ListAll(ctx, nil); err != nil {
				log.Fatalf("failed to get all apps: %s", err)
			} else {
				log.Println("getting envvars for each app...")
				for ix, app := range allApps {
					if envVars, err := cfClient.Applications.GetEnvironmentVariables(ctx, app.GUID); err != nil {
						fmt.Printf("%d - failed to get envvars for app %s: %s\n", ix, app.Name, err)
					} else {
						printString := ""
						labels2add := map[string]*string{}
						for _, envVar := range labels2Set {
							if value, ok := envVars[envVar]; ok {
								labels2add[envVar] = value
								printString = fmt.Sprintf("%s%s:%-15s ", printString, envVar, *envVars[envVar])
							} else {
								printString = fmt.Sprintf("%s%s:%-15s ", printString, envVar, "-")
							}
						}

						if len(labels2add) > 0 {
							if _, err = cfClient.Applications.Update(ctx, app.GUID, &resource.AppUpdate{Name: app.Name, Metadata: &resource.Metadata{Labels: labels2add}}); err != nil {
								fmt.Printf("%-4d failed to update app (guid %s) with splunk labels, error: %s\n", ix, app.GUID, err)
							} else {
								fmt.Printf("%-4d added labels to app (%s) %-50s: %s\n", ix, app.GUID, app.Name, printString)
							}
						} else {
							fmt.Printf("%-4d no labels for app   (%s) %-50s\n", ix, app.GUID, app.Name)
						}
					}
				}
			}
		}
	}
}
