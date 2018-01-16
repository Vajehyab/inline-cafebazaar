package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// TOKEN = "50758.QF5ZUPBRq2MlH3doOBcVV6IcPV8JiQW8w27qSIii"
	TOKEN = "50758.DtUOJsqB4dBIxhfnJWGfYRbjHKXV3LrGfmvZOVAS"

	// BASE_URL is API URL for vajehyab.
	// You can use "mina" before sending request to vajehyab.com
	// https://github.com/sariina/mina
	BASE_URL = "http://api.vajehyab.com/v3"
)

type VajehyabResponse struct {
	Response struct {
		Status bool `json:"status"`
		Code   int  `json:"code"`
	} `json:"response"`
	Meta struct {
		Q      string `json:"q"`
		Type   string `json:"type"`
		Filter string `json:"filter"`
	} `json:"meta"`
	Data struct {
		NumFound int `json:"num_found"`
		Results  []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			TitleEn string `json:"title_en"`
			Text    string `json:"text"`
			Source  string `json:"source"`
			Db      string `json:"db"`
			Num     int    `json:"num"`
		} `json:"results"`
	} `json:"data"`
}

type VajehyabSuggest struct {
	Response struct {
		Status bool `json:"status"`
		Code   int  `json:"code"`
	} `json:"response"`
	Meta struct {
		Q string `json:"q"`
	} `json:"meta"`
	Data struct {
		Suggestion []string `json:"suggestion"`
	} `json:"data"`
}

func sendRequest(search, dictionaries string) (VajehyabResponse, error) {
	result := VajehyabResponse{}

	client := &http.Client{}

	query := fmt.Sprintf("%s/search?token=%s&q=%s&type=exact&filter=%s", BASE_URL, TOKEN, search, dictionaries)
	req, err := http.NewRequest("GET", query, nil)

	if err != nil {
		return result, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return result, err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &result)

	return result, err
}

func getSuggestions(text string) (VajehyabSuggest, error) {
	result := VajehyabSuggest{}

	client := &http.Client{}

	query := fmt.Sprintf("%s/suggest?token=%s&q=%s", BASE_URL, TOKEN, text)
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return result, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return result, err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &result)

	return result, err
}
