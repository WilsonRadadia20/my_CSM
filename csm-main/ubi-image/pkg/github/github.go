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

func FetchDataGithub(gitUrl string) (Values, string, string, error) {

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
		return Values{}, "", "", errors
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

	return Values{gitTagVersion, gitImage, gitShaValue}, gitComment, gitFetchedData, nil

}
