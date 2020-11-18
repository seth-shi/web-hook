package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func isGitRepository(path string) error {

	output, err := shellExec("git rev-parse --is-inside-work-tree", path)
	if err != nil {
		return errors.New(fmt.Sprintf("%s is not git repository: %s", path, err.Error()))
	}
	if !strings.Contains(output, "true") {
		return errors.New(fmt.Sprintf("%s is not git repository: check %s", path, output))
	}

	return nil
}

func gitPull(path string) error {

	// pull code
	output, err := shellExec("git pull", path)
	if err != nil {
		return errors.New(fmt.Sprintf("git pull fail: %s", err.Error()))
	}
	if strings.Contains(output, "Already") {
		return errors.New(fmt.Sprintf("git Already"))
	}

	return nil
}

func getLastCommitId(path string) (string, error) {

	// pull code
	output, err := shellExec(`git log --pretty=format:"%H" -n 1 2>&1`, path)
	if err != nil {
		return "", errors.New(fmt.Sprintf("get commit id: %s", err.Error()))
	}

	reg := regexp.MustCompile(`^[0-9a-f]{40}$`)
	matches := reg.FindStringSubmatch(strings.Trim(string(output), "\""))
	if len(matches) == 0 {
		return "", errors.New("commit id len not match")
	}

	return matches[0], nil
}

func gitReset(hash, path string) error {
	// pull code ,
	_, err := shellExec("git reset --hard "+hash, path)

	return err
}

func shellExec(command, dir string) (string, error) {

	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}
	cmd := exec.Command(shell,flag,command)
	cmd.Dir = dir

	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
