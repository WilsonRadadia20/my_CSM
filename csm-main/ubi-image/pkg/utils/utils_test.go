package utils

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadYaml(t *testing.T) {
	t.Run("Config YAML Data Fetch: Testing", func(t *testing.T) {

		data := ConfigData{
			Urls: ConfigUrls{
				RedhatUrl: "https://redhat.abc",
				GithubUrl: "https://github.abc",
			},
			AuthData: ConfigGithubAuth{
				Owner:  "User123",
				Repo:   "my_repo",
				Token:  "xxxxxxxxxx",
				Path:   "path/to/file.mk",
				Branch: "branch1",
			},
			Comments: ConfigComment{
				Comment: "comment",
			},
			Reviewers: ConfigReviewers{
				Reviewers: "reviewer1,reviewer2,reviewer3",
			},
		}

		wordPtr := flag.String("f", "../../config_test.yaml", "config file")
		flag.Parse()
		config, err := ReadUtilConfigYaml(wordPtr)
		assert.Equal(t, config.Urls.RedhatUrl, data.Urls.RedhatUrl)
		assert.Equal(t, config.Urls.GithubUrl, data.Urls.GithubUrl)
		assert.Equal(t, config.Comments.Comment, data.Comments.Comment)
		assert.Equal(t, config.AuthData.Owner, data.AuthData.Owner)
		assert.Equal(t, config.AuthData.Repo, data.AuthData.Repo)
		assert.Equal(t, config.AuthData.Token, data.AuthData.Token)
		assert.Equal(t, config.AuthData.Path, data.AuthData.Path)
		assert.Equal(t, config.AuthData.Branch, data.AuthData.Branch)
		assert.Nil(t, err)

	})
}
