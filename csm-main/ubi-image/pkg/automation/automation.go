package automation

import (
	"fmt"
	"strings"
	"ubi-image/pkg/github"
	"ubi-image/pkg/redhat"
	"ubi-image/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type Values struct {
	TagVersion string
	Image      string
	Digests    string
}

type GithubValues struct {
	FetchedData       github.Values
	GithubFetchedData string
	GithubComment     string
}

type FetchedValues struct {
	GithubValues
	RedhatValues redhat.Values
}

func ComparingFetchedValues(fetchedValuesInstance FetchedValues) (comparisionResult bool) {
	//Trimming to exclude blank spaces
	if strings.TrimSpace(fetchedValuesInstance.RedhatValues.TagVersion) == strings.TrimSpace(fetchedValuesInstance.GithubValues.FetchedData.TagVersion) && strings.TrimSpace(fetchedValuesInstance.RedhatValues.Image) == strings.TrimSpace(fetchedValuesInstance.GithubValues.FetchedData.Image) && strings.TrimSpace(fetchedValuesInstance.RedhatValues.Digests) == strings.TrimSpace(fetchedValuesInstance.GithubValues.FetchedData.Digests) {
		return true
	} else {
		return false
	}
}

func updateWords(mainString string, oldSubString string, newSubString string) (replacedString string) {
	replacedString = strings.Replace(mainString, oldSubString, newSubString, -1)
	return
}

func UpdateContent(fetchedValuesInstance FetchedValues, newComment string) (newString string) {
	// \n so that the empty space remains as it is
	newString = updateWords(fetchedValuesInstance.GithubFetchedData, fetchedValuesInstance.GithubValues.FetchedData.TagVersion, fetchedValuesInstance.RedhatValues.TagVersion+"\n") //fetchedValuesInstance.redhatValues.tagVersion
	newString = updateWords(newString, fetchedValuesInstance.GithubValues.FetchedData.Image, fetchedValuesInstance.RedhatValues.Image+"\n")
	newString = updateWords(newString, fetchedValuesInstance.GithubValues.FetchedData.Digests, fetchedValuesInstance.RedhatValues.Digests)

	//Removing quotation marks
	newComment = strings.Trim(newComment, `"`)
	newString = updateWords(newString, fetchedValuesInstance.GithubValues.GithubComment, newComment+"\n")

	return newString
}

var configFileData utils.ConfigData
var githubAuth github.GithubAuth

func ReadConfigYaml(wordPtr *string) error {

	configData, isYamlError := utils.ReadUtilConfigYaml(wordPtr)
	configFileData = configData

	githubAuth = github.GithubAuth{GithubOwner: configFileData.AuthData.Owner, GithubRepo: configFileData.AuthData.Repo, GithubPath: configFileData.AuthData.Path, GithubToken: configFileData.AuthData.Token, GithubBranch: configFileData.AuthData.Branch}
	return isYamlError
}

func Process() error {

	//Reading Redhat Data
	RedhatValuesInstance, isRedhatError := redhat.FetchDataRedhat(configFileData.Urls.RedhatUrl)
	if isRedhatError != nil {
		log.Errorln("Failed to retrieve the redhat values: ", isRedhatError)
		return isRedhatError
	}

	//Reading Github Repo Data
	githubDataInstance, gitComment, gitFetchedData, isGithubError := github.FetchDataGithub(configFileData.Urls.GithubUrl)
	if isGithubError != nil {
		log.Errorln("Failed to retrieve the github values: ", isGithubError)
		return isGithubError
	}

	GithubValuesInstance := GithubValues{
		FetchedData:       githubDataInstance,
		GithubComment:     gitComment,
		GithubFetchedData: gitFetchedData,
	}

	fetchedValuesInstance := FetchedValues{
		GithubValues: GithubValuesInstance,
		RedhatValues: RedhatValuesInstance,
	}

	//Error handling if the data is not retrieved
	if fetchedValuesInstance.RedhatValues.TagVersion == "" || fetchedValuesInstance.RedhatValues.Image == "" || fetchedValuesInstance.RedhatValues.Digests == "" {
		log.Errorln("Nothing Fetched!!!")
		isNothing := fmt.Errorf("nothing fetched")
		return isNothing
	} else if fetchedValuesInstance.GithubValues.FetchedData.TagVersion == "" || fetchedValuesInstance.GithubValues.FetchedData.Image == "" || fetchedValuesInstance.FetchedData.Digests == "" {
		log.Errorln("Nothing Fetched!!!")
		isNothing := fmt.Errorf("nothing fetched")
		return isNothing
	}

	log.Infoln("Red Hat Catalog data fetched")
	log.Infoln("Github Repo data fetched")

	//Comparining the Redhat and Github fetched values
	isResultSame := ComparingFetchedValues(fetchedValuesInstance)

	if isResultSame {
		log.Infoln("\nNothing to be changed")
		return nil
	}

	log.Infoln("There is new update")
	contentAfterChanges := UpdateContent(fetchedValuesInstance, configFileData.Comments.Comment)
	// log.Println(contentAfterChanges + "\n")

	isBranchVerifyError := githubAuth.GitVerifyBranch()
	if isBranchVerifyError != nil {
		log.Errorln("Error verifying branch", isBranchVerifyError)
		return isBranchVerifyError
	}

	//Creating new branch in github
	isBranchError := githubAuth.CreateGitBranch()
	if isBranchError != nil {
		log.Errorln("Error creating branch", isBranchError)
		return isBranchError
	}

	data := &github.ContentToChange{Content: contentAfterChanges}
	// log.Println(data)

	// Git Push in branch
	isPushError := githubAuth.GithubPush(data)
	if isPushError != nil {
		log.Errorln("Error updating file content", isPushError)
		return isPushError
	}

	//Git PR
	isPullError := githubAuth.GithubPullRequest(data, fetchedValuesInstance.RedhatValues.TagVersion)
	if isPullError != nil {
		log.Errorln("Error creating pull request:", isPullError)
		return isPullError
	}

	//Add Reviewers
	isAddReviwerError := githubAuth.GithubPrAddReviewers(configFileData.Reviewers.Reviewers)
	if isAddReviwerError != nil {
		log.Errorln("Error adding reviewers to Pull request:", isAddReviwerError)
		return isAddReviwerError
	}

	return nil
}
