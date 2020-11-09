package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

const (
	HookFilterParameter = "parameters"
	HookFilterHeader    = "header"
)

type GitHook struct {
	Name          string         `json:"name"`
	Dir           string         `json:"dir"`
	Notifications []Notification `json:"notifications"`
	HookFilters   []HookFilter   `json:"hook_filters"`
	Hooks         []Hook         `json:"hooks"`
	FailHooks     []FailHook     `json:"fail_hooks"`
}

type Notification struct {
	Type    string `json:"type"`
	WebHook string `json:"web_hook"`
}

type FailHook struct {
	Dir   string `json:"dir"`
	Shell string `json:"shell"`
}

type Hook struct {
	Dir                string `json:"dir"`
	Shell              string `json:"shell"`
	Assert             string `json:"assert"`
	AssertNo           string `json:"assert_no"`
	AssertFailContinue bool   `json:"assert_fail_continue"`
}

type HookFilter struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
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
