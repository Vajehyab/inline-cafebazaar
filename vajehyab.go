package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	TOKEN    = "50758.QF5ZUPBRq2MlH3doOBcVV6IcPV8JiQW8w27qSIii"
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
