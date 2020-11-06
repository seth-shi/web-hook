package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)


func isGitRepository(path string) error {

	output, err := shellExec("git rev-parse --is-inside-work-tree", path)
	if err != nil {
		return errors.New(fmt.Sprintf("%s is not git repository: %s", path, err.Error()))
	}
	if ! strings.Contains(output, "true") {
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
	output, err := shellExec("git log --pretty=format:\"%H\" -n 1 2>&1", path)
	if err != nil {
		return "", errors.New(fmt.Sprintf("get commit id: %s", err.Error()))
	}

	hash := strings.TrimSpace(string(output))
	if len(output) != 40 {
		return "", errors.New("commit id len not match")
	}

	return hash, nil
}

func gitReset(hash, path string) error {
	// pull code ,
	_, err := shellExec("git reset --hard " + hash, path)

	return err
}


func shellExec(command, dir string) (string, error) {

	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}