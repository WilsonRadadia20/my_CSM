package github

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestGithub(t *testing.T) {
	t.Run("Github Data Fetch: Testing", func(t *testing.T) {

		chromeInstance, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		chromeInstance, cancel = chromedp.NewContext(chromeInstance)
		defer cancel()

		githubUrl, err := filepath.Abs("../testPages/githubTestPage.mhtml")
		assert.Nil(t, err)

		gitcomment := "# Common base image for all CSM images. When base image is upgraded, update the following 3 lines with URL, Version, and DEFAULT_BASEIMAGE variable.\n"

		v := Values{
			TagVersion: "ubi9/ubi-micro 9.3-13\n",
			Image:      "https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=65a8f97db7e4bede96526c22\n",
			Digests:    "registry.access.redhat.com/ubi9/ubi-micro@sha256:d72202acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d",
		}

		fetchedData1 := "# Common base image for all CSM images. When base image is upgraded, update the following 3 lines with URL, Version, and DEFAULT_BASEIMAGE variable.\n"
		fetchedData2 := "# URL: https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=65a8f97db7e4bede96526c22\n"
		fetchedData3 := "# Version: ubi9/ubi-micro 9.3-13"
		fetchedData4 := "\nDEFAULT_BASEIMAGE=\"registry.access.redhat.com/ubi9/ubi-micro@sha256:d72202acf3073b61cb407e86395935b7bac5b93b16071d2b40b9fb485db2135d\""
		fetchedData5 := "\nDEFAULT_GOIMAGE=\"golang:1.22\""

		fetchedData := fetchedData1 + fetchedData2 + fetchedData3 + fetchedData4 + fetchedData5

		gitTagVersion, gitImage, gitShaValue, gitComment, gitFetchedData, err := GithubDataScrap(chromeInstance, githubUrl)

		assert.Equal(t, v.TagVersion, gitTagVersion)
		assert.Equal(t, v.Image, gitImage)
		assert.Equal(t, v.Digests, gitShaValue)
		assert.Equal(t, gitcomment, gitComment)
		assert.Equal(t, fetchedData, gitFetchedData)
		assert.Nil(t, err)
	})
}
