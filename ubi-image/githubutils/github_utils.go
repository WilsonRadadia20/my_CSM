package githubutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

var prLink string

type GithubAuth struct {
	GithubOwner  string
	GithubRepo   string
	GithubPath   string
	GithubToken  string
	GithubBranch string
}

type ContentToChange struct {
	Content string
}

func initializeContext(githubAuth GithubAuth) (ctx context.Context, client *github.Client) {
	ctx = context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAuth.GithubToken})

	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)
	return
}

func extractPRNumber(prUrl string) (string, error) {
	u, err := url.Parse(prUrl)
	if err != nil {
		return "", err
	}
	tokens := strings.Split(u.Path, "/")

	prNumber := tokens[len(tokens)-1]
	return prNumber, nil
}

func (githubAuth GithubAuth) GitVerifyBranch() {

	ctx, client := initializeContext(githubAuth)

	_, _, err := client.Git.GetRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/"+githubAuth.GithubBranch)
	if err != nil {
		log.Println("The branch does not exist", err)
	} else {
		log.Println("The branch exist, deleting the branch")
		_, errors := client.Git.DeleteRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/"+githubAuth.GithubBranch)
		if errors != nil {
			log.Println("The branch does not exist", errors)
		}
	}
}

func (githubAuth GithubAuth) CreateGitBranch() error {

	ctx, client := initializeContext(githubAuth)

	ref, _, err := client.Git.GetRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/main")
	if err != nil {
		return err
	}

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/" + githubAuth.GithubBranch),
		Object: &github.GitObject{SHA: ref.Object.SHA},
	}

	_, _, errors := client.Git.CreateRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, newRef)
	if errors != nil {
		return errors
	}

	log.Println("New branch created successfully")

	return nil
}

func (githubAuth GithubAuth) GithubPush(data *ContentToChange) error {

	ctx, client := initializeContext(githubAuth)

	//Fetching the content of the file specified in the path and storing in fileContent
	fileContent, _, _, err := client.Repositories.GetContents(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, githubAuth.GithubPath, &github.RepositoryContentGetOptions{Ref: githubAuth.GithubBranch})
	if err != nil {
		return err
	}
	sha := fileContent.GetSHA()

	commitMessage := "Update new version"

	// Modifying the file by the new content
	opts := &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: []byte(data.Content),
		SHA:     &sha,
		Branch:  &githubAuth.GithubBranch,
	}

	// Update the file
	_, _, errors := client.Repositories.UpdateFile(context.Background(), githubAuth.GithubOwner, githubAuth.GithubRepo, githubAuth.GithubPath, opts)
	if errors != nil {
		return errors
	}

	log.Println("File updated successfully")

	return nil
}

func (githubAuth GithubAuth) GithubPullRequest(data *ContentToChange, tagVersion string) error {

	ctx, client := initializeContext(githubAuth)

	baseBranch := "main"
	headBranch := githubAuth.GithubBranch

	subject := tagVersion + " " + "version update"

	//Crateing PR
	pr, _, err := client.PullRequests.Create(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, &github.NewPullRequest{
		Title:               github.String(subject),
		Head:                github.String(headBranch),
		Base:                github.String(baseBranch),
		Body:                github.String("Updation in the version of tag, image and sha value"),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		return err
	}

	log.Printf("Pull request created successfully")
	prLink = *pr.HTMLURL
	log.Println("Pull request link:", *pr.HTMLURL)

	return nil
}

func (githubAuth GithubAuth) GithubPrAddReviewers(reviewersString string) error {

	prNum, errLink := extractPRNumber(prLink)
	if errLink != nil {
		fmt.Println("Error: ", errLink)
		return errLink
	}

	reviewers := strings.Split(reviewersString, ",")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers", githubAuth.GithubOwner, githubAuth.GithubRepo, prNum)

	payload := map[string][]string{"reviewers": reviewers}
	payloadByte, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		log.Println("Error marshaling JSON:", errMarshal)
		return errMarshal
	}

	req, errRequest := http.NewRequest("POST", url, bytes.NewReader(payloadByte))
	if errRequest != nil {
		fmt.Println("Error creating request: ", errRequest)
		return errRequest
	}

	req.Header.Set("Content-Type", "application/json")

	req.SetBasicAuth("", githubAuth.GithubToken)

	client := &http.Client{}

	resp, errResponse := client.Do(req)
	if errResponse != nil {
		log.Println("Error sending request:", errResponse)
		return errResponse
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		log.Println("Reviewers added successfully")

	} else {
		log.Println("Error adding reviewer. Status code:", resp.StatusCode)

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error is there", err)
			return err
		}

		log.Println("Response Body:", string(responseBody))
		return err
	}

	return nil
}

func githubutils() {

}
