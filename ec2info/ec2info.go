package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var MetaTagName = "ccc_EC2MetadataException"
var devAccount = "282694560246"
var prdAccount = "701895972238"
var profileName = "default"

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("please specify the file name to parse...")
		os.Exit(8)
	}
	filename := os.Args[1]
	//log.Printf("reading file %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("failed reading input file %s: %s\n", filename, err)
		os.Exit(8)
	}
	//log.Printf("json-parsing file %s", filename)
	var ec2info EC2Info
	err = json.Unmarshal(file, &ec2info)
	if err != nil {
		log.Printf("failed unmarshalling json from file %s, error: %s\n", filename, err)
		os.Exit(8)
	}

	var totalInstances = 0
	var cmdLines []string
	var namePart1, namePart2 string
	for _, reservation := range ec2info.Reservations {
		//fmt.Printf("reservation %d ID: %s\n", ix, reservation.ReservationID)
		for _, instance := range reservation.Instances {
			if instance.State.Name != "terminated" {
				metadataTag := fmt.Sprintf("(%s tag missing)", MetaTagName)
				instanceName := ""
				director := ""
				pcfenv := ""
				for _, tag := range instance.Tags {
					if tag.Key == MetaTagName {
						metadataTag = ""
					}
					if tag.Key == "director" {
						director = tag.Value
					}
					if tag.Key == "pcfenv" || tag.Key == "cfenv" {
						pcfenv = tag.Value
					}
					if tag.Key == "Name" {
						instanceName = tag.Value
						parts := strings.Split(instanceName, "/")
						namePart1 = parts[0]
						if len(parts) > 1 {
							namePart2 = parts[1]
						}
					}
				}

				if reservation.OwnerID == prdAccount {
					profileName = "pim"
				}
				if instance.MetadataOptions.HTTPTokens == "required" {
					cmdLines = append(cmdLines, fmt.Sprintf("aws --profile %s ec2 modify-instance-metadata-options --instance-id %s --http-tokens=optional", profileName, instance.InstanceID))
				}
				fmt.Printf("  %s %3s %9s %-30s %37s HttpTokens:%s %s\n", instance.InstanceID, pcfenv, director, namePart1, namePart2, instance.MetadataOptions.HTTPTokens, metadataTag)
				totalInstances++
			}
		}
	}
	fmt.Printf("\nTotal running instances: %d\n", totalInstances)
	if len(cmdLines) > 0 {
		fmt.Print("issue following commands to correct instances:\n\n")
		for _, cmdLine := range cmdLines {
			fmt.Println(cmdLine)
		}
	}
}
