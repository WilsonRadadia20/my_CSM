package automation

import (
	"flag"
	"path/filepath"
	"testing"
	"ubi-image/pkg/github"
	"ubi-image/pkg/redhat"
	"ubi-image/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestBadAutomation(t *testing.T) {

	var testbadGithubAuth github.GithubAuth
	var mockconfigFileDataTest utils.ConfigData
	var RedhatValuesInstance redhat.Values

	wordPtrTest := flag.String("f", "../../config_test.yaml", "config test file")

	t.Run("Read Yaml: Testing", func(t *testing.T) {
		mockConfigDataTest, isReadConfigYamlErrorTest := utils.ReadUtilConfigYaml(wordPtrTest)
		mockconfigFileDataTest = mockConfigDataTest
		testbadGithubAuth = github.GithubAuth{GithubOwner: mockconfigFileDataTest.AuthData.Owner, GithubRepo: mockconfigFileDataTest.AuthData.Repo, GithubPath: mockconfigFileDataTest.AuthData.Path, GithubToken: mockconfigFileDataTest.AuthData.Token, GithubBranch: mockconfigFileDataTest.AuthData.Branch}
		assert.Nil(t, isReadConfigYamlErrorTest)
		assert.NotNil(t, mockconfigFileDataTest)
	})
	t.Run("Fetch Redhat Data: Bad Testing", func(t *testing.T) {
		redhatUrl, err := filepath.Abs(mockconfigFileDataTest.Urls.RedhatUrl)
		redhatInstance, isRedhatError := redhat.FetchDataRedhat(redhatUrl)
		RedhatValuesInstance = redhatInstance
		assert.Nil(t, err)
		assert.Nil(t, isRedhatError)
		assert.NotNil(t, RedhatValuesInstance.Digests)
		assert.NotNil(t, RedhatValuesInstance.Image)
		assert.NotNil(t, RedhatValuesInstance.TagVersion)
	})
	var githubDataInstance github.Values
	var mockGitComment string
	var mockGitFetchedData string
	t.Run("Fetch Github Data: Bad Testing", func(t *testing.T) {
		githubUrl, err := filepath.Abs(mockconfigFileDataTest.Urls.GithubUrl)
		githubDataInstance, mockGitComment, mockGitFetchedData, isGithubError := github.FetchDataGithub(githubUrl)
		// isGithubError = errors.New("Error in fetching data")
		assert.Nil(t, err)
		assert.Nil(t, isGithubError)
		assert.NotNil(t, githubDataInstance.Digests)
		assert.NotNil(t, githubDataInstance.Image)
		assert.NotNil(t, githubDataInstance.TagVersion)
		assert.NotNil(t, mockGitComment)
		assert.NotNil(t, mockGitFetchedData)
	})
	t.Run("Github Branch Verification: Bad Testing", func(t *testing.T) {
		isBranchVerifyError := testbadGithubAuth.GitVerifyBranch()
		assert.Nil(t, isBranchVerifyError)
	})
	t.Run("Github Create Branch: Bad Testing", func(t *testing.T) {
		isCreateBranchError := testbadGithubAuth.CreateGitBranch()
		assert.NotNil(t, isCreateBranchError)
	})

	GithubValuesInstance := GithubValues{
		FetchedData:       githubDataInstance,
		GithubComment:     mockGitComment,
		GithubFetchedData: mockGitFetchedData,
	}

	fetchedValuesInstance := FetchedValues{
		GithubValues: GithubValuesInstance,
		RedhatValues: RedhatValuesInstance,
	}
	t.Run("Comparing github and redhat fetched values: Bad Testing", func(t *testing.T) {
		isResultSame := ComparingFetchedValues(fetchedValuesInstance)
		assert.False(t, isResultSame)
	})
	var contentAfterChanges string
	t.Run("Updating content", func(t *testing.T) {
		contentAfterChanges = UpdateContent(fetchedValuesInstance, mockconfigFileDataTest.Comments.Comment)
		assert.NotNil(t, contentAfterChanges)
	})

	data := &github.ContentToChange{Content: contentAfterChanges}
	t.Run("Github Push: Bad Testing", func(t *testing.T) {
		isGitPushError := testbadGithubAuth.GithubPush(data)
		assert.NotNil(t, isGitPushError)
	})
	t.Run("Github Pull Request: Bad Testing", func(t *testing.T) {
		isGitPullRequestError := testbadGithubAuth.GithubPullRequest(data, fetchedValuesInstance.RedhatValues.TagVersion)
		assert.NotNil(t, isGitPullRequestError)
	})
	t.Run("Github PR Add Reviews: Bad Testing", func(t *testing.T) {
		isGitAddReviewersError := testbadGithubAuth.GithubPrAddReviewers(mockconfigFileDataTest.Reviewers.Reviewers)
		assert.Nil(t, isGitAddReviewersError)
	})

}
