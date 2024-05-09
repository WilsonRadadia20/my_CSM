package utils

import (
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type ConfigData struct {
	Urls      ConfigUrls       `yaml:"urls"`
	AuthData  ConfigGithubAuth `yaml:"githubAuth"`
	Comments  ConfigComment    `yaml:"comments"`
	Reviewers ConfigReviewers  `yaml:"gihtubReviewers"`
}

type ConfigUrls struct {
	RedhatUrl string `yaml:"redhatUrl"`
	GithubUrl string `yaml:"githubUrl"`
}

type ConfigGithubAuth struct {
	Owner  string `yaml:"owner"`
	Repo   string `yaml:"repo"`
	Token  string `yaml:"token"`
	Path   string `yaml:"path"`
	Branch string `yaml:"branch"`
}

type ConfigComment struct {
	Comment string `yaml:"comment"`
}

type ConfigReviewers struct {
	Reviewers string `yaml:"reviewers"`
}

func ReadUtilConfigYaml(wordPtr *string) (ConfigData, error) {
	var config ConfigData

	//Reading file
	yamlFile, isReadFileErr := os.ReadFile(*wordPtr)

	if isReadFileErr != nil {
		return ConfigData{}, isReadFileErr
	}

	//Decoding data
	//Unmarshal: First parameter is byte slice and second parameter is pointer to struct
	isUnmarshalError := yaml.Unmarshal(yamlFile, &config)
	if isUnmarshalError != nil {
		return ConfigData{}, isUnmarshalError
	}
	log.Infoln("Yaml file data fetched")
	return config, nil
}
