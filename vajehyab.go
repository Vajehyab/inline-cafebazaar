package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	TOKEN = "50758.QF5ZUPBRq2MlH3doOBcVV6IcPV8JiQW8w27qSIii"

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

func getWord(search, dictionaries string) (VajehyabResponse, error) {
	result := VajehyabResponse{}
	query := fmt.Sprintf("%s/search?token=%s&q=%s&type=exact&filter=%s", BASE_URL, TOKEN, search, dictionaries)
	err := getJSON(query, &result)
	return result, err
}

func getSuggestion(text string) (VajehyabSuggest, error) {
	result := VajehyabSuggest{}
	query := fmt.Sprintf("%s/suggest?token=%s&q=%s", BASE_URL, TOKEN, text)
	err := getJSON(query, &result)
	return result, err
}

func getJSON(url string, output interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(respBody, &output)
}

func encodeDictionary(dic map[string]bool) string {
	output := ""

	for _, key := range keys {
		if dic[key] {
			output += key + ","
		}
	}
	output = output[:len(output)-1]

	return output
}
