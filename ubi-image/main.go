package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"
	L "ubi-image/githubutils"

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
	githubComment     string
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

func comparingFetchedValues(fetchedValuesInstance fetchedValues) (comparisionResult bool) {

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

func updateContent(fetchedValuesInstance fetchedValues, newComment string) (newString string) {
	// Testing data:
	// images1 := "https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=56b9f97db7e4bede96526c22\n"
	// tags := "ubi9/ubi-micro 9.3-15\n"
	// sha := "registry.access.redhat.com/ubi9/ubi-micro@sha256:b88902acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d"

	// \n so that the empty space remains as it is
	newString = updateWords(fetchedValuesInstance.githubFetchedData, fetchedValuesInstance.githubValues.fetchedData.tagVersion, fetchedValuesInstance.redhatValues.tagVersion+"\n") //fetchedValuesInstance.redhatValues.tagVersion
	newString = updateWords(newString, fetchedValuesInstance.githubValues.fetchedData.image, fetchedValuesInstance.redhatValues.image+"\n")
	newString = updateWords(newString, fetchedValuesInstance.githubValues.fetchedData.digests, fetchedValuesInstance.redhatValues.digests)

	//Removing quotation marks
	newComment = strings.Trim(newComment, `"`)
	newString = updateWords(newString, fetchedValuesInstance.githubValues.githubComment, newComment+"\n")

	return newString
}

func readConfigYaml(wordPtr *string) (error, ConfigData) {
	var config ConfigData

	//Reading file
	yamlFile, err := os.ReadFile(*wordPtr)

	if err != nil {
		return err, ConfigData{}
	}

	//Decoding data
	//Unmarshal: First parameter is byte slice and second parameter is pointer to struct
	errors := yaml.Unmarshal(yamlFile, &config)
	if errors != nil {
		return errors, ConfigData{}
	}
	log.Println("Yaml file data fetched")
	return nil, config
}

func fetchDataRedhat(redhatUrl string) (values, error) {

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
	var tagVersion string

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
		return values{}, err
	}

	tagVersion = repository + " " + tagVersion

	return values{tagVersion, url, manifestList}, nil
}

func fetchDataGithub(gitUrl string) (values, string, string, error) {

	var gitFetchedData string
	var gitComment string
	var gitTagVersion string
	var gitImage string
	var gitShaValue string

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
		return values{}, "", "", errors
	}

	//Extracting comment
	startIndexComment := strings.Index(gitFetchedData, "# Common")
	endIndexComment := strings.Index(gitFetchedData, "# URL")
	gitComment = gitFetchedData[startIndexComment:endIndexComment]

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

	return values{gitTagVersion, gitImage, gitShaValue}, gitComment, gitFetchedData, nil

}

func main() {
	//for command line
	wordPtr := flag.String("f", "config/config.yaml", "config file")

	flag.Parse()

	//reading YAML file
	isYamlError, configFileData := readConfigYaml(wordPtr)
	if isYamlError != nil {
		log.Println("Error reading the yaml file", isYamlError)
		return
	}

	var githubAuth = L.GithubAuth{GithubOwner: configFileData.AuthData.Owner, GithubRepo: configFileData.AuthData.Repo, GithubPath: configFileData.AuthData.Path, GithubToken: configFileData.AuthData.Token, GithubBranch: configFileData.AuthData.Branch}

	//Reading Redhat Data
	redhatValuesInstance, isRedhatError := fetchDataRedhat(configFileData.Urls.RedhatUrl)
	if isRedhatError != nil {
		log.Println("Failed to retrieve the redhat values: ", isRedhatError)
		return
	}

	//Reading Github Repo Data
	githubDataInstance, gitComment, gitFetchedData, isGithubError := fetchDataGithub(configFileData.Urls.GithubUrl)
	if isGithubError != nil {
		log.Println("Failed to retrieve the github values: ", isGithubError)
		return
	}

	githubValuesInstance := githubValues{
		fetchedData:       githubDataInstance,
		githubComment:     gitComment,
		githubFetchedData: gitFetchedData,
	}

	fetchedValuesInstance := fetchedValues{
		githubValues: githubValuesInstance,
		redhatValues: redhatValuesInstance,
	}

	//Error handling if the data is not retrieved
	if fetchedValuesInstance.redhatValues.tagVersion == "" || fetchedValuesInstance.redhatValues.image == "" || fetchedValuesInstance.redhatValues.digests == "" {
		log.Println("Nothing Fetched!!!")
		return
	} else if fetchedValuesInstance.githubValues.fetchedData.tagVersion == "" || fetchedValuesInstance.githubValues.fetchedData.image == "" || fetchedValuesInstance.fetchedData.digests == "" {
		log.Println("Nothing Fetched!!!")
		return
	}

	//Printing the retrieved data
	// log.Println("The data retrieved from Red Hat Catalog")
	// log.Println("The latest tag version is:", fetchedValuesInstance.redhatValues.tagVersion)
	// log.Println("The image value is:", fetchedValuesInstance.redhatValues.image)
	// log.Println("The sha value is:", fetchedValuesInstance.redhatValues.digests)

	// log.Println("\nThe data retrieved from Github Repo")
	// log.Print("The latest tag version is: ", fetchedValuesInstance.githubValues.fetchedData.tagVersion)
	// log.Print("The image value is: ", fetchedValuesInstance.githubValues.fetchedData.image)
	// log.Print("The sha value is: ", fetchedValuesInstance.githubValues.fetchedData.digests)

	log.Println("Red Hat Catalog data fetched")
	log.Println("Github Repo data fetched")

	//Comparining the Redhat and Github fetched values
	isResultSame := comparingFetchedValues(fetchedValuesInstance)

	if isResultSame {
		log.Println("\nNothing to be changed")
	} else {

		log.Println("There is new update")
		contentAfterChanges := updateContent(fetchedValuesInstance, configFileData.Comments.Comment)
		// log.Println(contentAfterChanges + "\n")

		githubAuth.GitVerifyBranch()

		//Creating new branch in github
		isBranchError := githubAuth.CreateGitBranch()
		if isBranchError != nil {
			log.Println("Error creating branch", isBranchError)
			return
		}

		data := &L.ContentToChange{Content: contentAfterChanges}
		// log.Println(data)

		// Git Push in branch
		isPushError := githubAuth.GithubPush(data)
		if isPushError != nil {
			log.Println("Error updating file content", isPushError)
			return
		}

		//Git PR
		isPullError := githubAuth.GithubPullRequest(data, fetchedValuesInstance.redhatValues.tagVersion)
		if isPullError != nil {
			log.Println("Error creating pull request:", isPullError)
			return
		}

		//Add Reviewers
		isAddReviwerError := githubAuth.GithubPrAddReviewers(configFileData.Reviewers.Reviewers)
		if isAddReviwerError != nil {
			log.Println("Error adding reviewers to Pull request:", isAddReviwerError)
			return
		}
	}
}
