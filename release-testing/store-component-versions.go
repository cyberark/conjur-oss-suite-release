package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
)

func main() {
	var suiteFile string
	flag.StringVar(&suiteFile, "f", "", "Repository YAML file to parse")
	flag.Parse()

	if suiteFile == "" {
		flag.PrintDefaults()
		log.Fatal("missing required -f flag")
		return
	}

	repoConfig, err := repositories.NewConfig(suiteFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, categories := range repoConfig.Section.Categories {
		for _, repo := range categories.Repos {
			key := fmt.Sprintf("release.%s.version", repo.Name)
			log.Println("Setting store value for key: " + key)

			err = storeClient.Set(
				key,
				strings.TrimPrefix(repo.Version, "v"),
			)

			if err != nil {
				log.Fatal(err)
				return
			}
		}
	}
}
