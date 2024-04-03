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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: org-spaces-info <directory containing spaceConfig.yml files>")
		os.Exit(1)
	}
	directory := os.Args[1]
	fmt.Printf("Directory: %s\n", directory)

	if err := filepath.Walk(directory, showPath); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Total spaces: %d, sukkels: %d\n", totalCount, sukkelCount)
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
				samlUsers := len(spaceConfig.SpaceDeveloper.SamlUsers) + len(spaceConfig.SpaceAuditor.SamlUsers) + len(spaceConfig.SpaceManager.SamlUsers)
				aadGroups := len(spaceConfig.SpaceDeveloper.AadGroups) + len(spaceConfig.SpaceAuditor.AadGroups) + len(spaceConfig.SpaceManager.AadGroups)
				if samlUsers > 0 {
					sukkelCount++
					fmt.Printf("%s/%s: saml_users:%d, aad_groups:%d\n", spaceConfig.Org, spaceConfig.Space, samlUsers, aadGroups)
				}
			}
		}
	}
	return nil
}
