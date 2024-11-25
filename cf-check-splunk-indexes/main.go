package main

import (
	"bufio"
	"context"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"log"
	"os"
	"strings"
)

var splunkIndexes = make(map[string]int)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <cf_home_dir> <location of allowed-splunk-indexes-file> (i.e. ~/workspace/panzer-config-prd/splunk_indexes/azure_prd.json)", os.Args[0])
	}
	// read a yml file with allowed splunk indexes
	// Open the file
	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", os.Args[2], err)
	}
	defer func() { _ = file.Close() }()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), ":")
		if len(words) > 1 {
			indexName := strings.TrimSpace(words[0])
			indexName = strings.Trim(indexName, `"`)
			splunkIndexes[indexName] = 0
		}
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
				log.Println("getting envvars and checking SPLUNK_INDEX for each app...")
				for ix, app := range allApps {
					if envVars, err := cfClient.Applications.GetEnvironmentVariables(ctx, app.GUID); err != nil {
						log.Printf("%d - failed to get envvars for app %s: %s\n", ix, app.Name, err)
					} else {
						if value, ok1 := envVars["SPLUNK_INDEX"]; ok1 {
							if count, ok2 := splunkIndexes[*value]; ok2 {
								log.Printf("%-4d app %s - %-64s state=%s has   valid SPLUNK_INDEX %s\n", ix, app.GUID, app.Name, app.State, *value)
								splunkIndexes[*value] = count + 1
							} else {
								log.Printf("%-4d app %s - %-64s state=%s has invalid SPLUNK_INDEX %s\n", ix, app.GUID, app.Name, app.State, *value)
							}
						}
					}
				}
				for k, v := range splunkIndexes {
					log.Printf("index usage for %-30s : %4d\n", k, v)
				}
			}
		}
	}
}
