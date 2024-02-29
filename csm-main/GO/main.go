package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

type githubFetchedValues struct {
	githubTagVersion  string
	githubImageValue  string
	githubShaValue    string
	githubFetchedData string
}

type redhatFetchedValues struct {
	redhatTagVersion string
	redhatImageValue string
	redhatShaValue   string
}

type fetchedValues struct {
	githubFetchedValues
	redhatFetchedValues
}

type githubAuth struct {
	githubOwner  string
	githubRepo   string
	githubPath   string
	githubToken  string
	githubBranch string
}

func compareFetchedValues(fetchedValues fetchedValues) (comparisionResult bool) {

	//Trimming to exclude blank spaces
	if strings.TrimSpace(fetchedValues.redhatTagVersion) == strings.TrimSpace(fetchedValues.githubTagVersion) && strings.TrimSpace(fetchedValues.redhatImageValue) == strings.TrimSpace(fetchedValues.githubImageValue) && strings.TrimSpace(fetchedValues.redhatShaValue) == strings.TrimSpace(fetchedValues.githubShaValue) {
		return true
	} else {
		return false
	}
}

func updateWords(mainString string, oldSubString string, newSubString string) (replacedString string) {
	replacedString = strings.Replace(mainString, oldSubString, newSubString, -1)
	return
}

func updateContent(fetchedValues fetchedValues) (newString string) {
	// Testing data:
	// images1 := "56b9f97db7e4bede96526c22\n"
	// tags := "9.3-11\n"
	// sha := "b88902acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d"

	// \n so that the empty space remains as it is
	newString = updateWords(fetchedValues.githubFetchedData, fetchedValues.githubImageValue, "56b9f97db7e4bede96526c22 "+"\n")
	newString = updateWords(newString, fetchedValues.githubTagVersion, "9.3-12"+"\n")
	newString = updateWords(newString, fetchedValues.githubShaValue, "b88902acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d")

	return newString
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

	err := chromedp.Run(ctx,
		//Navigating to the website
		chromedp.Navigate(redhatUrl),

		//Wait until main is visible
		chromedp.WaitVisible(`main`, chromedp.ByQuery),

		//Retrieving the value of Tag
		chromedp.Text(`main span.eco-static-tag span.eco-static-tag__name`, &tagVersion, chromedp.NodeVisible, chromedp.ByQuery),

		//Retrieving the image
		//chromedp.Sleep(5*time.Second), //The image is in website's url so to get image website must be loaded completely.
		chromedp.Location(&url),

		//Retrieving the sha value main div.pf-c-description-list__group dd.pf-c-description-list__description div.pf-c-clipboard-copy__group input#text-input-45
		chromedp.Evaluate(`document.querySelectorAll("main span.pf-c-tabs__item-text")[document.querySelectorAll("main span.pf-c-tabs__item-text").length - 1].click();`, nil),
		chromedp.Sleep(5*time.Second), //For loading the website after the button[Get this image] is clicked
		chromedp.Evaluate(`document.querySelector("input.pf-c-form-control").value;`, &manifestList),
	)

	//Error handling
	if err != nil {
		fmt.Println("Failed to retrieve the redhat values: ", err)
		return "", "", ""
	}

	//Splittion to get image and sha
	imageValueArr := strings.Split(url, "image=")
	shaValueArr := strings.Split(manifestList, "sha256:")

	return tagVersion, imageValueArr[1], shaValueArr[1]

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
		return "", "", "", ""
	}

	//Extracting tag version
	startIndexTag := strings.Index(gitFetchedData, "ubi-micro ")
	endIndexTag := strings.Index(gitFetchedData, "DEFAULT_BASEIMAGE=")
	gitTagVersion = gitFetchedData[startIndexTag+10 : endIndexTag] //+10 to exclude ubi-micro(10 char)

	//Extracting image value
	startIndexImage := strings.Index(gitFetchedData, "image=")
	endIndexImage := strings.Index(gitFetchedData, "# Version:")
	gitImage = gitFetchedData[startIndexImage+6 : endIndexImage] //+6 to exclude image=(6 char)

	//Extracting sha value
	startIndexSha := strings.Index(gitFetchedData, "sha256:")
	endIndexSha := strings.Index(gitFetchedData, "DEFAULT_GOIMAGE")
	gitShaValue = gitFetchedData[startIndexSha+7 : endIndexSha-2] //+7 to exclude sha256:(7 char) and -2 to exclude _"(2 char)

	return gitTagVersion, gitImage, gitShaValue, gitFetchedData

}

func githubPush(contentAfterChanges string, githubAuth githubAuth) {

	ctx := context.Background()

	//Creating static token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAuth.githubToken})

	//Passing the token and context for creating HTTP client
	tc := oauth2.NewClient(ctx, ts)

	// Create a GitHub client using the access token
	client := github.NewClient(tc)

	// Specification
	owner := githubAuth.githubOwner
	repo := githubAuth.githubRepo
	branch := githubAuth.githubBranch
	path := githubAuth.githubPath
	newContent := contentAfterChanges
	commitMessage := "Update new version"

	//Fetching the content of the file specified in the path and storing in fileContent
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		log.Fatal(err)
	}
	sha := fileContent.GetSHA()

	// Modifying the file by the new content
	opts := &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: []byte(newContent),
		SHA:     &sha,
		Branch:  &branch,
	}

	// Update the file
	_, _, err = client.Repositories.UpdateFile(context.Background(), owner, repo, path, opts)
	if err != nil {
		log.Fatal("Error updating file content:", err)
	}

	fmt.Println("File updated successfully!")

}

func githubPullRequest(contentAfterChanges string, githubAuth githubAuth) {

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAuth.githubToken})

	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	//Specifications
	owner := githubAuth.githubOwner
	repo := githubAuth.githubRepo
	baseBranch := "main"
	headBranch := githubAuth.githubBranch

	//Crateing PR
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String("Updation in version"),
		Head:                github.String(headBranch),
		Base:                github.String(baseBranch),
		Body:                github.String("Updation in the version of tag, image and sha value"),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		fmt.Println("Error creating pull request:", err)
		os.Exit(1)
	}

	fmt.Printf("Pull request created: %s\n", *pr.HTMLURL)
}

func main() {

	redhatUrl := `https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64`
	gitUrl := `https://github.com/WilsonRadadia20/my_CSM/blob/main/csm-main/config/csm-common.mk`

	// var githubAuth = githubAuth{"WilsonRadadia20", "my_CSM", "csm-main/config/csm-common.mk", "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", "User_branch_1"}

	tagVersion, imageValue, shaValue := fetchDataRedhat(redhatUrl)
	gitTagVersion, gitImageValue, gitShaValue, gitFetchedData := fetchDataGithub(gitUrl)

	var fetchedValues = fetchedValues{
		githubFetchedValues: githubFetchedValues{gitTagVersion, gitImageValue, gitShaValue, gitFetchedData},
		redhatFetchedValues: redhatFetchedValues{tagVersion, imageValue, shaValue}}

	//Error handling if the data is not retrieved //
	if fetchedValues.redhatTagVersion == "" && fetchedValues.redhatImageValue == "" && fetchedValues.redhatShaValue == "" {
		// fmt.Println("Nothing Fetched!!!")
		return
	} else if fetchedValues.githubTagVersion == "" && fetchedValues.githubImageValue == "" && fetchedValues.githubShaValue == "" {
		// fmt.Println("Nothing Fetched!!!")
		return
	}

	//Printing the retrieved data
	fmt.Println("The data retrieved from Red Hat Catalog")
	fmt.Println("The latest tag version is:", fetchedValues.redhatTagVersion)
	fmt.Println("The image value is:", fetchedValues.redhatImageValue)
	fmt.Println("The sha value is:", fetchedValues.redhatShaValue)

	fmt.Println("\nThe data retrieved from github repo")
	fmt.Print("The latest tag version is: ", fetchedValues.githubTagVersion)
	fmt.Print("The image value is: ", fetchedValues.githubImageValue)
	fmt.Print("The sha value is: ", fetchedValues.githubShaValue)

	comparisionResult := compareFetchedValues(fetchedValues)

	if comparisionResult {
		fmt.Println("\nNothing to be changed")
	} else {

		fmt.Println("\nThere is new update\n\n")
		contentAfterChanges := updateContent(fetchedValues)
		fmt.Println(contentAfterChanges)

		//Git Push in branch
		// githubPush(contentAfterChanges, githubAuth)

		//Git PR
		// githubPullRequest(contentAfterChanges, githubAuth)
	}

}
