package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type GitHook struct {
	Name          string         `json:"name"`
	Dir           string         `json:"dir"`
	Notifications []Notification `json:"notifications"`
	Hooks         []Hook         `json:"hooks"`
}

type Notification struct {
	Type    string `json:"type"`
	WebHook string `json:"web_hook"`
}

type Hook struct {
	Shell              string `json:"shell"`
	Assert             string `json:"assert"`
	AssertFailContinue bool   `json:"assert_fail_continue"`
}

func parseHooks(file string) (map[string]GitHook, error) {

	hookBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// parse all git repository
	var repositories []GitHook
	err = json.Unmarshal(hookBytes, &repositories)
	if err != nil {
		return nil, err
	}

	if len(repositories) == 0 {
		return nil, errors.New("repositories len is 0")
	}

	repositoryMap := make(map[string]GitHook, len(repositories))
	for _, repository := range repositories {
		repositoryMap[repository.Name] = repository
	}

	return repositoryMap, nil
}
