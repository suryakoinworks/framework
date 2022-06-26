package parsers

import (
	"bytes"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const ModulePath = "configs/modules.yaml"

type module struct {
	Config []string `yaml:"modules"`
}

func ParseModule(dir string) []string {
	var path bytes.Buffer
	path.WriteString(dir)
	path.WriteString("/")
	path.WriteString(ModulePath)

	config, err := os.ReadFile(path.String())
	mapping := module{}
	if err != nil {
		log.Println(err)

		return []string{}
	}

	err = yaml.Unmarshal(config, &mapping)
	if err != nil {
		log.Println(err)

		return []string{}
	}

	return mapping.Config
}
