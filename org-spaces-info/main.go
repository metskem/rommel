package main

import (
	"bufio"
	"fmt"
	"github.com/metskem/rommel/org-spaces-info/model"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var totalCount, sukkelCount int
var recipients map[string]bool

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: org-spaces-info <directory containing spaceConfig.yml files>")
		os.Exit(1)
	}
	directory := os.Args[1]
	//fmt.Printf("Directory: %s\n", directory)
	recipients = make(map[string]bool)

	if err := filepath.Walk(directory, showPath); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Total spaces: %d, sukkels: %d\n\n", totalCount, sukkelCount)
	for k := range recipients {
		_, _ = fmt.Fprintln(os.Stderr, k+";")
	}
}

func showPath(fullPath string, info os.FileInfo, err error) error {
	if err == nil && info.Name() == "spaceConfig.yml" {
		var file *os.File
		if file, err = os.Open(fullPath); err != nil {
			fmt.Printf("%s could not be opened: %s\n", fullPath, err)
			return nil
		} else {
			defer func() { _ = file.Close() }()
			decoder := yaml.NewDecoder(bufio.NewReader(file))
			decoder.KnownFields(true)
			spaceConfig := model.SpaceConfig{}
			if err = decoder.Decode(&spaceConfig); err != nil {
				fmt.Printf("%s could not be parsed: %s\n", fullPath, err)
			} else {
				totalCount++
				var valuesToPrint []string
				for _, value := range spaceConfig.SpaceDeveloper.Users {
					if value != "cf-dsmdev-automation" && value != "cf-dsmprd-automation" && value != "pcf-panzer" {
						valuesToPrint = append(valuesToPrint, value)
						recipients[spaceConfig.Metadata.Contact] = true
					}
				}

				if len(valuesToPrint) > 0 {
					sukkelCount++
					fmt.Printf("%s/%s: %v\n", spaceConfig.Org, spaceConfig.Space, valuesToPrint)
				}
			}
		}
	}
	return nil
}
