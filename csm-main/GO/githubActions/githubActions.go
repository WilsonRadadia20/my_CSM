package githubActions

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

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

func (githubAuth GithubAuth) GitVerifyBranch() {

	ctx, client := initializeContext(githubAuth)

	_, _, err := client.Git.GetRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/"+githubAuth.GithubBranch)
	if err != nil {
		fmt.Println("The branch does not exist", err)
		log.Fatal("The branch does not exist", err)
	} else {
		fmt.Println("The branch exist, deleting the branch")
		_, errors := client.Git.DeleteRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/"+githubAuth.GithubBranch)
		if errors != nil {
			fmt.Println("The branch does not exist", errors)
			log.Fatal("The branch does not exist", errors)
			return
		}
	}
}

func (githubAuth GithubAuth) CreateGitBranch() {
	ctx, client := initializeContext(githubAuth)

	ref, _, err := client.Git.GetRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, "refs/heads/main")
	if err != nil {
		fmt.Println("Error creating branch", err)
		log.Fatal("Error creating branch", err)
		return
	}

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/" + githubAuth.GithubBranch),
		Object: &github.GitObject{SHA: ref.Object.SHA},
	}

	_, _, errors := client.Git.CreateRef(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, newRef)
	if errors != nil {
		fmt.Println("Error creating branch", errors)
		log.Fatal("Error creating branch", errors)
		return
	}

	fmt.Println("New branch created successfully")
}

func (githubAuth GithubAuth) GithubPush(data *ContentToChange) {

	ctx, client := initializeContext(githubAuth)

	//Fetching the content of the file specified in the path and storing in fileContent
	fileContent, _, _, err := client.Repositories.GetContents(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, githubAuth.GithubPath, &github.RepositoryContentGetOptions{Ref: githubAuth.GithubBranch})
	if err != nil {
		fmt.Println("Error updating file content", err)
		log.Fatal("Error updating file content:", err)
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
		fmt.Println("Error updating file content", errors)
		log.Fatal("Error updating file content:", errors)
		return
	}

	fmt.Println("File updated successfully!")

}

func (githubAuth GithubAuth) GithubPullRequest(data *ContentToChange) {

	ctx, client := initializeContext(githubAuth)

	baseBranch := "main"
	headBranch := githubAuth.GithubBranch

	//Crateing PR
	pr, _, err := client.PullRequests.Create(ctx, githubAuth.GithubOwner, githubAuth.GithubRepo, &github.NewPullRequest{
		Title:               github.String("Updation in version"),
		Head:                github.String(headBranch),
		Base:                github.String(baseBranch),
		Body:                github.String("Updation in the version of tag, image and sha value"),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		log.Fatal("Error creating pull request", err)
		fmt.Println("Error creating pull request:", err)
		return
	}

	fmt.Printf("Pull request created: %s\n", *pr.HTMLURL)
}

func githubActions() {

}
