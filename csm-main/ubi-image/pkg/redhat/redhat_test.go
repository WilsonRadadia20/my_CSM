package redhat

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestRedhat(t *testing.T) {
	t.Run("Redhat Data Fetch: Testing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ctx, cancel = chromedp.NewContext(ctx)
		defer cancel()

		redhatUrl, err := filepath.Abs("../testPages/redhatTestPage.html")

		v := Values{
			TagVersion: "ubi9/ubi-micro 9.4-6",
			Image:      "https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=662a8edd22c80ead7411ec6c",
			Digests:    "registry.access.redhat.com/ubi9/ubi-micro@sha256:826cf6250899228070dcd4eb0abd8667d0468a5fe0148d54bb513c912b06cee4",
		}
		tagversion, image, digest, err := RedhatDataScrap(ctx, redhatUrl)
		assert.Equal(t, tagversion, v.TagVersion)
		assert.NotNil(t, image)
		assert.Equal(t, digest, v.Digests)
		assert.Nil(t, err)
	})
}
