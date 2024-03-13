package main

import (
	L "GO/githubActions"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v3"
)

type values struct {
	tagVersion string
	image      string
	digests    string
}

type githubValues struct {
	fetchedData       values
	githubFetchedData string
}

type fetchedValues struct {
	githubValues
	redhatValues values
}

type githubAuth struct {
	githubOwner  string
	githubRepo   string
	githubPath   string
	githubToken  string
	githubBranch string
}

type ConfigData struct {
	Urls     ConfigUrls       `yaml:"urls"`
	AuthData ConfigGithubAuth `yaml:"githubAuth"`
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

func compareFetchedValues(fetchedValuesInstance fetchedValues) (comparisionResult bool) {

	//Trimming to exclude blank spaces
	if strings.TrimSpace(fetchedValuesInstance.redhatValues.tagVersion) == strings.TrimSpace(fetchedValuesInstance.githubValues.fetchedData.tagVersion) && strings.TrimSpace(fetchedValuesInstance.redhatValues.image) == strings.TrimSpace(fetchedValuesInstance.githubValues.fetchedData.image) && strings.TrimSpace(fetchedValuesInstance.redhatValues.digests) == strings.TrimSpace(fetchedValuesInstance.githubValues.fetchedData.digests) {
		return true
	} else {
		return false
	}
}

func updateWords(mainString string, oldSubString string, newSubString string) (replacedString string) {
	replacedString = strings.Replace(mainString, oldSubString, newSubString, -1)
	return
}

func updateContent(fetchedValuesInstance fetchedValues) (newString string) {
	// Testing data:
	// images1 := "https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=56b9f97db7e4bede96526c22\n"
	// tags := "ubi9/ubi-micro 9.3-15\n"
	// sha := "registry.access.redhat.com/ubi9/ubi-micro@sha256:b88902acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d"

	// \n so that the empty space remains as it is
	newString = updateWords(fetchedValuesInstance.githubFetchedData, fetchedValuesInstance.githubValues.fetchedData.tagVersion, fetchedValuesInstance.redhatValues.tagVersion+"\n") //fetchedValuesInstance.redhatValues.tagVersion
	newString = updateWords(newString, fetchedValuesInstance.githubValues.fetchedData.image, fetchedValuesInstance.redhatValues.image+"\n")
	newString = updateWords(newString, fetchedValuesInstance.githubValues.fetchedData.digests, fetchedValuesInstance.redhatValues.digests)

	return newString
}

func readConfigYaml(config ConfigData) ConfigData {

	//Reading file
	yamlFile, err := ioutil.ReadFile("config/config.yaml")

	if err != nil {
		fmt.Println("Error reading the yaml file", err)
		log.Fatal("Error reading the yaml file", err)
	}

	//Decoding data
	//Unmarshal: First parameter is byte slice and second parameter is pointer to struct
	errors := yaml.Unmarshal(yamlFile, &config)
	if errors != nil {
		fmt.Println("Error reading the yaml file", errors)
		log.Fatal("Error reading the yaml file", errors)
	}

	return config
}

func fetchDataRedhat(redhatUrl string) (tagVersion string, imageValue string, shaValue string) {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize a controllable Chrome instance by using empty context
	ctx, cancel = chromedp.NewContext(ctx)

	//To release the browser resources when it is no longer needed[Scope]
	defer cancel()

	//Scraping Logic
	var url string
	var manifestList string
	var repository string

	err := chromedp.Run(ctx,
		//Navigating to the website
		chromedp.Navigate(redhatUrl),

		//Wait until main is visible
		chromedp.WaitVisible(`main`, chromedp.ByQuery),

		//Retrieving the value of Tag
		chromedp.Text(`main span.eco-static-tag span.eco-static-tag__name`, &tagVersion, chromedp.NodeVisible, chromedp.ByQuery),

		//Retrieving repository
		chromedp.Text(`main div.eco-container-repo--registry`, &repository, chromedp.NodeVisible, chromedp.ByQuery),

		//Retrieving the image
		chromedp.Sleep(5*time.Second), //The image is in website's url so to get image website must be loaded completely.
		chromedp.Location(&url),

		//Retrieving the sha value main div.pf-c-description-list__group dd.pf-c-description-list__description div.pf-c-clipboard-copy__group input#text-input-45
		chromedp.Evaluate(`document.querySelectorAll("main span.pf-c-tabs__item-text")[document.querySelectorAll("main span.pf-c-tabs__item-text").length - 1].click();`, nil),
		// chromedp.Sleep(5*time.Second), //For loading the website after the button[Get this image] is clicked
		chromedp.Evaluate(`document.querySelectorAll("input.pf-c-form-control")[4].value;`, &manifestList),
	)

	//Error handling
	if err != nil {
		fmt.Println("Failed to retrieve the redhat values: ", err)
		log.Fatal("Failed to retrieve the redhat values: ", err)
		return "", "", ""
	}

	tagVersion = repository + " " + tagVersion

	return tagVersion, url, manifestList

}

func fetchDataGithub(gitUrl string) (gitTagVersion string, gitImage string, gitShaValue string, gitFetchedData string) {

	// Create a context with a timeout
	chromeInstance, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initializing chrome instance
	chromeInstance, cancel = chromedp.NewContext(chromeInstance)
	defer cancel()

	errors := chromedp.Run(chromeInstance,
		chromedp.Navigate(gitUrl),
		chromedp.Value(`textarea#read-only-cursor-text-area`, &gitFetchedData, chromedp.ByQuery),
	)

	//Error handling
	if errors != nil {
		fmt.Println("Failed to retrieve the github values: ", errors)
		log.Fatal("Failed to retrieve the redhat values: ", errors)
		return "", "", "", ""
	}

	//Extracting tag version
	startIndexTag := strings.Index(gitFetchedData, "# Version: ")
	endIndexTag := strings.Index(gitFetchedData, "DEFAULT_BASEIMAGE=")
	gitTagVersion = gitFetchedData[startIndexTag+11 : endIndexTag] //+10 to exclude ubi-micro(10 char)

	//Extracting image value
	startIndexImage := strings.Index(gitFetchedData, "# URL: ")
	endIndexImage := strings.Index(gitFetchedData, "# Version:")
	gitImage = gitFetchedData[startIndexImage+7 : endIndexImage] //+6 to exclude image=(6 char)

	//Extracting sha value
	startIndexSha := strings.Index(gitFetchedData, "DEFAULT_BASEIMAGE=\"")
	endIndexSha := strings.Index(gitFetchedData, "DEFAULT_GOIMAGE")
	gitShaValue = gitFetchedData[startIndexSha+19 : endIndexSha-2] //+7 to exclude sha256:(7 char) and -2 to exclude _"(2 char)

	return gitTagVersion, gitImage, gitShaValue, gitFetchedData

}

func main() {

	var config ConfigData
	configFileData := readConfigYaml(config)

	var githubAuth = L.GithubAuth{configFileData.AuthData.Owner, configFileData.AuthData.Repo, configFileData.AuthData.Path, configFileData.AuthData.Token, configFileData.AuthData.Branch}

	tagVersion, imageValue, shaValue := fetchDataRedhat(configFileData.Urls.RedhatUrl)
	gitTagVersion, gitImageValue, gitShaValue, gitFetchedData := fetchDataGithub(configFileData.Urls.GithubUrl)

	redhatDataInstance := values{tagVersion, imageValue, shaValue}

	githubValuesInstance := githubValues{
		fetchedData:       values{gitTagVersion, gitImageValue, gitShaValue},
		githubFetchedData: gitFetchedData,
	}

	fetchedValuesInstance := fetchedValues{
		githubValues: githubValuesInstance,
		redhatValues: redhatDataInstance,
	}

	//Error handling if the data is not retrieved
	if fetchedValuesInstance.redhatValues.tagVersion == "" || fetchedValuesInstance.redhatValues.image == "" || fetchedValuesInstance.redhatValues.digests == "" {
		fmt.Println("Nothing Fetched!!!")
		log.Fatal("Nothing Fetched!!!")
		return
	} else if fetchedValuesInstance.githubValues.fetchedData.tagVersion == "" || fetchedValuesInstance.githubValues.fetchedData.image == "" || fetchedValuesInstance.fetchedData.digests == "" {
		fmt.Println("Nothing Fetched!!!")
		log.Fatal("Nothing Fetched!!!")
		return
	}

	//Printing the retrieved data
	fmt.Println("The data retrieved from Red Hat Catalog")
	fmt.Println("The latest tag version is:", fetchedValuesInstance.redhatValues.tagVersion)
	fmt.Println("The image value is:", fetchedValuesInstance.redhatValues.image)
	fmt.Println("The sha value is:", fetchedValuesInstance.redhatValues.digests)

	fmt.Println("\nThe data retrieved from Github Repo")
	fmt.Print("The latest tag version is: ", fetchedValuesInstance.githubValues.fetchedData.tagVersion)
	fmt.Print("The image value is: ", fetchedValuesInstance.githubValues.fetchedData.image)
	fmt.Print("The sha value is: ", fetchedValuesInstance.githubValues.fetchedData.digests)

	comparisionResult := compareFetchedValues(fetchedValuesInstance)

	if comparisionResult {
		fmt.Println("\nNothing to be changed")
		log.Fatal("Nothing to be changed")
	} else {

		fmt.Println("\nThere is new update\n")
		contentAfterChanges := updateContent(fetchedValuesInstance)
		// log.Fatal(contentAfterChanges)
		// fmt.Println(contentAfterChanges + "\n")

		githubAuth.GitVerifyBranch()

		//Creating new branch in github
		githubAuth.CreateGitBranch()

		data := &L.ContentToChange{Content: contentAfterChanges}

		//Git Push in branch
		githubAuth.GithubPush(data)

		//Git PR
		githubAuth.GithubPullRequest(data)
	}

}
