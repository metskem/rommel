package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-uaa"
	"log"
	"os"
	"time"
)

var (
	apiD04, apiD05, apiD06, apiP05, apiP06, apiP07, apiP08 *uaa.API
	err                                                    error
)

func main() {
	log.SetOutput(os.Stderr)
	if apiD04, err = uaa.New("https://uaa.sys.cfd04.aws.rabo.cloud", uaa.WithClientCredentials("admin", "bsPUlfm38fbCOWsc8T39YaeMDbJiS4", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiD05, err = uaa.New("https://uaa.sys.cfd05.rabobank.nl", uaa.WithClientCredentials("admin", "QrLPaEyL2hYgMPB2eQDkmA3ljgamxy", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiD06, err = uaa.New("https://uaa.sys.cfd06.rabobank.nl", uaa.WithClientCredentials("admin", "NMJcFi8cenyypKRQj3hAX7PFe1pgCZ", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiP05, err = uaa.New("https://uaa.sys.cfp05.aws.rabo.cloud", uaa.WithClientCredentials("admin", "U1xVvPnNhwT2ruEAM8Tn4feKWp5i7B", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiP06, err = uaa.New("https://uaa.sys.cfp06.aws.rabo.cloud", uaa.WithClientCredentials("admin", "nWjbTO0NiJ7sTIzhy7QiTtr1yBcUIA", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiP07, err = uaa.New("https://uaa.sys.cfp07.rabobank.nl", uaa.WithClientCredentials("admin", "ygovvFuJt0mgGZGgADwMi8JN2S3e7S", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	if apiP08, err = uaa.New("https://uaa.sys.cfp08.rabobank.nl", uaa.WithClientCredentials("admin", "W7X8Ng2RSx3plQRGGW5IBe1a56xusj", uaa.JSONWebToken)); err != nil {
		log.Fatal(err)
	}
	allApis := make(map[string]*uaa.API)

	// switch between the set for DEV and the set for PRD

	//allApis["d04"] = apiD04
	//allApis["d05"] = apiD05
	allApis["d06"] = apiD06

	//allApis["p05"] = apiP05
	//allApis["p06"] = apiP06
	//allApis["p07"] = apiP07
	//allApis["p08"] = apiP08

	for _, api := range allApis {
		log.Printf("Fetching all users from %s...", api.TargetURL)
		if users, err := api.ListAllUsers("", "id", "", uaa.SortAscending); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Fetched %d users\n", len(users))
			for _, user := range users {
				//if user.LastLogonTime == 0 && user.Name.FamilyName == "" {
				fmt.Printf("Created: %s, LastLogonTime: %s, Origin: %-22s, Guid: %s, FamilyName: %s, Username: %s\n", user.Meta.Created, time.UnixMilli(int64(user.LastLogonTime)).Format(time.RFC3339), user.Origin, user.ID, user.Name.FamilyName, user.Username)
				//}
			}
		}
	}
}
