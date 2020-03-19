package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
)

func main() {
	var suiteFile string
	flag.StringVar(&suiteFile, "f", "", "Repository YAML file to parse")
	flag.Parse()

	if suiteFile == "" {
		flag.PrintDefaults()
		log.ErrLogger.Fatal("missing required -f flag")
		return
	}

	repoConfig, err := repositories.NewConfig(suiteFile)
	if err != nil {
		log.ErrLogger.Fatal(err)
		return
	}

	for _, categories := range repoConfig.Section.Categories {
		for _, repo := range categories.Repos {
			key := fmt.Sprintf("RELEASE.%s.VERSION", strings.ToUpper(repo.Name))
			log.OutLogger.Println("Setting store value for key: " + key)

			err = storeClient.Set(
				key,
				strings.TrimPrefix(repo.Version, "v"),
			)

			if err != nil {
				log.ErrLogger.Fatal(err)
				return
			}
		}
	}
}
