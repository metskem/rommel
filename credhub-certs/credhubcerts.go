package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("please specify the file name to parse...\n")
		os.Exit(8)
	}
	filename := os.Args[1]
	fmt.Printf("reading file %s\n", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("failed reading input file %s: %s\n", filename, err)
		os.Exit(8)
	}
	//log.Printf("json-parsing file %s", filename)
	var certlist CertificateList
	if err = json.Unmarshal(file, &certlist); err != nil {
		fmt.Printf("failed to parse file %s: %s\n", filename, err)
		os.Exit(8)
	} else {
		fmt.Printf("found %d certificates\n\n", len(certlist.Certificates))
		for _, certificate := range certlist.Certificates {
			fmt.Printf("%s:\n", certificate.Name)
			for _, version := range certificate.Versions {
				fmt.Printf("  %s - %s\n", version.ExpiryDate, version.ID)
			}
		}
	}

}

type CertificateList struct {
	Certificates []struct {
		ID       string        `json:"id"`
		Name     string        `json:"name"`
		SignedBy string        `json:"signed_by"`
		Signs    []interface{} `json:"signs"`
		Versions []struct {
			CertificateAuthority bool      `json:"certificate_authority"`
			ExpiryDate           time.Time `json:"expiry_date"`
			Generated            bool      `json:"generated"`
			ID                   string    `json:"id"`
			SelfSigned           bool      `json:"self_signed"`
			Transitional         bool      `json:"transitional"`
		} `json:"versions"`
	} `json:"certificates"`
}
