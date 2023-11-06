package api

import (
	"fmt"
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"net/http"
)

func DoRequest(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	request.Header.Set("Authorization", fmt.Sprintf(""))

	response, err := client.Do(request)
	return response, err
}

func Get(uri string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", config.VATUSA_API2_URL, uri)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(req)
}
