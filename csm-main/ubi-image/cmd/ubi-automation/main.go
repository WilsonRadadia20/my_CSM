package main

import (
	"flag"
	"ubi-image/pkg/automation"

	log "github.com/sirupsen/logrus"
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

func main() {
	//for command line
	wordPtr := flag.String("f", "config.yaml", "config file")

	flag.Parse()

	//reading YAML file
	isYamlError := automation.ReadConfigYaml(wordPtr) //from utils
	if isYamlError != nil {
		log.Errorln("Error reading the yaml file", isYamlError)
		return
	}

	isProcessError := automation.Process()
	if isProcessError != nil {
		log.Errorln("Error in the automation process", isProcessError)
		return
	}

}
