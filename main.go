package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	"github.com/royeo/dingrobot"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	log             *logging.Logger
	logFile         *os.File
	hookChan        chan string
	gitRepositories map[string]GitHook
)

func init() {

	var err error

	logFile, err = openLogFile("logs/app.log")
	if err != nil {
		panic(err)
	}

	log, err = getLogger(logFile)
	if err != nil {
		panic(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gitRepositories, err = parseHooks("hooks.json")
	if err != nil {
		log.Fatal(err)
	}

	hookChan = make(chan string, len(gitRepositories)*5)
}

func main() {

	defer logFile.Close()

	go handleHooks()

	router := gin.Default()
	router.GET("/ping", ping)
	router.GET("/hooks/:name", webHook)
	router.POST("/hooks/:name", webHook)

	err := router.Run(fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatal(err)
	}
}

func ping(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func webHook(c *gin.Context) {

	name := c.Param("name")

	hookChan <- name

	c.String(http.StatusOK, "hook %s success", name)
}

func handleHooks() {

	for name := range hookChan {

		handleHook(name)
	}
}

func handleHook(name string) error {

	var err error
	var buildOutput []string

	repository, exists := gitRepositories[name]
	if ! exists {
		return errors.New(fmt.Sprintf("repository [%s] does not exist\n", name))
	}

	defer func() {

		if err := recover(); err != nil {

			log.Error(err)
			for _, notification := range repository.Notifications {
				sendNotification(notification, buildOutput, err)
			}
		}
	}()

	g, err := git.PlainOpen(repository.Dir)
	if err != nil {
		panic(err)
	}

	w, err := g.Worktree()
	if err != nil {
		panic(err)
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		panic(err)
	}

	// exec all shell
	for _, hook := range repository.Hooks {

		output, err := shellExec(hook.Shell)
		if err != nil {
			panic(err)
		}

		buildOutput = append(buildOutput, output)
		if ! strings.Contains(output, hook.Assert) {
			if hook.AssertFailContinue {
				continue
			}

			panic(fmt.Sprintf("[output] %s \n[assert] %s\n", output, hook.Assert))
		}
	}

	return nil
}

func shellExec(command string) (string, error) {

	cmd := exec.Command("/bin/bash", "-c", command)
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func sendNotification(n Notification, buildOutput []string, err interface{})  {

	if n.Type == "dingtalk" {

		robot := dingrobot.NewRobot(n.WebHook)
		body := "- build steps \n"
		for _, o := range buildOutput {
			body += fmt.Sprintf("- %s\n", o)
		}

		if errObj, ok := err.(error); ok {
			body += fmt.Sprintf("### err: %s", errObj.Error())
		}

		err := robot.SendMarkdown("build notification", body, nil, true)
		if err != nil {
			log.Error(err)
		}
	}
}