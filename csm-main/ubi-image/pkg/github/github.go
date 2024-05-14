package github

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Values struct {
	TagVersion string
	Image      string
	Digests    string
}

func DataExtraction(gitFetchedData string, startIndexData string, endIndexData string, startindex int, endIndex int) string {
	startIndexComment := strings.Index(gitFetchedData, startIndexData)
	endIndexComment := strings.Index(gitFetchedData, endIndexData)
	return gitFetchedData[startIndexComment+startindex : endIndexComment-endIndex]
}

func GithubDataScrap(chromeInstance context.Context, gitUrl string) (string, string, string, string, string, error) {

	var gitFetchedData string

	errors := chromedp.Run(chromeInstance,
		chromedp.Navigate(gitUrl),
		chromedp.Value(`textarea#read-only-cursor-text-area`, &gitFetchedData, chromedp.ByQuery),
	)

	//Error handling
	if errors != nil {
		return "", "", "", "", "", errors
	}

	//Extracting comment
	gitComment := DataExtraction(gitFetchedData, "# Common", "# URL", 0, 0)

	//Extracting tag version
	gitTagVersion := DataExtraction(gitFetchedData, "# Version: ", "DEFAULT_BASEIMAGE=", 11, 0) //+10 to exclude ubi-micro(10 char)

	//Extracting image value
	gitImage := DataExtraction(gitFetchedData, "# URL: ", "# Version:", 7, 0) //+6 to exclude image=(6 char)

	//Extracting sha value
	gitShaValue := DataExtraction(gitFetchedData, "DEFAULT_BASEIMAGE=\"", "DEFAULT_GOIMAGE", 19, 2) //+7 to exclude sha256:(7 char) and -2 to exclude _"(2 char)

	return gitTagVersion, gitImage, gitShaValue, gitComment, gitFetchedData, nil
}

func FetchDataGithub(gitUrl string) (Values, string, string, error) {

	// Create a context with a timeout
	chromeInstance, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initializing chrome instance
	chromeInstance, cancel = chromedp.NewContext(chromeInstance)
	defer cancel()

	gitTagVersion, gitImage, gitShaValue, gitComment, gitFetchedData, isGithubDataScrapError := GithubDataScrap(chromeInstance, gitUrl)
	if isGithubDataScrapError != nil {
		return Values{}, "", "", isGithubDataScrapError
	}

	return Values{gitTagVersion, gitImage, gitShaValue}, gitComment, gitFetchedData, nil

}
