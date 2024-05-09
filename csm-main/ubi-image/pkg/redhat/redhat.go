package redhat

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Values struct {
	TagVersion string
	Image      string
	Digests    string
}

func FetchDataRedhat(redhatUrl string) (Values, error) {

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
	var repository string
	var tagVersion string

	err := chromedp.Run(ctx,
		//Navigating to the website
		chromedp.Navigate(redhatUrl),

		//Wait until main is visible
		chromedp.WaitVisible(`main`, chromedp.ByQuery),

		//Retrieving the value of Tag
		chromedp.Text(`main span.eco-static-tag span.eco-static-tag__name`, &tagVersion, chromedp.NodeVisible, chromedp.ByQuery),

		//Retrieving repository
		chromedp.Text(`main div.eco-container-repo--registry`, &repository, chromedp.NodeVisible, chromedp.ByQuery),

		//Retrieving the image
		chromedp.Sleep(5*time.Second), //The image is in website's url so to get image website must be loaded completely.
		chromedp.Location(&url),

		//Retrieving the sha value main div.pf-c-description-list__group dd.pf-c-description-list__description div.pf-c-clipboard-copy__group input#text-input-45
		chromedp.Evaluate(`document.querySelectorAll("main span.pf-c-tabs__item-text")[document.querySelectorAll("main span.pf-c-tabs__item-text").length - 1].click();`, nil),
		// chromedp.Sleep(5*time.Second), //For loading the website after the button[Get this image] is clicked
		chromedp.Evaluate(`document.querySelectorAll("input.pf-c-form-control")[4].value;`, &manifestList),
	)

	//Error handling
	if err != nil {
		return Values{}, err
	}

	tagVersion = repository + " " + tagVersion

	return Values{tagVersion, url, manifestList}, nil
}
