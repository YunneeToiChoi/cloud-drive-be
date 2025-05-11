package client

import (
	"common-lib/discovery"
	"io"
	"log"
	"net/http"
	"net/url"
)

func GetUserId(id string) ([]byte, error) {
	baseURL := discovery.BuildServiceURL("users", "GetById")
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	query.Add("userId", id)
	parsedURL.RawQuery = query.Encode()

	finalURL := parsedURL.String()

	log.Printf("Calling user service atttttttttt: %s", finalURL)

	res, err := http.Get(finalURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func CallUserServiceHealth() ([]byte, error) {
	url := discovery.BuildServiceURL("users", "health")

	log.Printf("Checking health at: %s", url)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
