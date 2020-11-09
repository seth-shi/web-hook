package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	"github.com/royeo/dingrobot"
	"net/http"
	"os"
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

	go task()

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

	repository, exists := gitRepositories[name]
	if !exists {
		c.String(http.StatusOK, "hook %s doesn't exists", name)
		return
	}

	if len(repository.HookFilters) > 0 {

		var requestData map[string]string
		_ = c.Bind(&requestData)

		for _, filter := range repository.HookFilters {

			switch filter.Type {

			case HookFilterHeader:
				if filter.Value != c.GetHeader(filter.Key) {
					c.String(http.StatusBadRequest, "header %s want %s, give %s", filter.Key, filter.Value, c.GetHeader(filter.Key))
					return
				}
			case HookFilterParameter:
				// support query, post,
				if filter.Value != requestData[filter.Key] {
					c.String(http.StatusBadRequest, "header %s want %s, give %s", filter.Key, filter.Value, requestData[filter.Key])
					return
				}
			}
		}
	}

	hookChan <- name

	c.String(http.StatusOK, "hook %s success", name)
}

func task() {

	for name := range hookChan {

		err := taskJob(name)
		if err != nil {
			log.Error(err)
		}
	}
}

func taskJob(name string) error {

	var err error
	var lastCommitId string
	var buildOutput []string

	repository, exists := gitRepositories[name]
	if !exists {
		return errors.New(fmt.Sprintf("repository [%s] does not exist\n", name))
	}

	defer func() {

		var err error
		if e := recover(); e != nil {
			log.Error(e)

			if str, ok := e.(string); ok {
				err = errors.New(str)
			}

			// rollback
			if len(lastCommitId) != 0 {

				re := gitReset(lastCommitId, repository.Dir)
				buildOutput = append(buildOutput, "git reset --hard "+lastCommitId, fmt.Sprintf("%s", re))
			}
		}

		for _, hook := range repository.FailHooks {
			o := handleFailHookShell(repository, hook)
			buildOutput = append(buildOutput, hook.Shell, o)
		}

		for _, notification := range repository.Notifications {
			sendNotification(notification, buildOutput, err)
		}
	}()

	if err = isGitRepository(repository.Dir); err != nil {
		panic(err.Error())
	}

	lastCommitId, err = getLastCommitId(repository.Dir)
	if err != nil {
		panic(err.Error())
	}

	if err = gitPull(repository.Dir); err != nil {
		panic(err.Error())
	}

	// exec all shell
	for _, hook := range repository.Hooks {

		o := handleHookShell(repository, hook)
		buildOutput = append(buildOutput, hook.Shell, o)
	}

	return nil
}

func handleHookShell(repository GitHook, hook Hook) string {

	dir := repository.Dir
	if len(hook.Dir) > 0 {
		dir = hook.Dir
	}

	output, err := shellExec(hook.Shell, dir)
	if err != nil {
		panic(fmt.Sprintf("exec [%s] fail :%s", hook.Shell, err.Error()))
	}

	if len(hook.Assert) > 0 && !strings.Contains(output, hook.Assert) {
		if !hook.AssertFailContinue {
			panic(fmt.Sprintf("[output] %s \n[assert] %s\n", output, hook.Assert))
		}

	}

	if len(hook.AssertNo) > 0 && strings.Contains(output, hook.AssertNo) {
		if !hook.AssertFailContinue {
			panic(fmt.Sprintf("[output] %s \n[assert_no] %s\n", output, hook.Assert))
		}
	}

	return output
}

func handleFailHookShell(repository GitHook, hook FailHook) string {

	dir := repository.Dir
	if len(hook.Dir) > 0 {
		dir = hook.Dir
	}

	output, err := shellExec(hook.Shell, dir)
	if err != nil {
		panic(fmt.Sprintf("exec [%s] fail :%s", hook.Shell, err.Error()))
	}

	return output
}

func sendNotification(n Notification, buildOutput []string, err error) {

	if n.Type == "dingtalk" {

		robot := dingrobot.NewRobot(n.WebHook)
		body := "-build steps \n"
		for _, o := range buildOutput {
			body += fmt.Sprintf("- %s \n", o)
		}

		title := "build success"
		if err != nil {
			title = "build fail"
			body += fmt.Sprintf("## err: %s", err.Error())
		}

		body = fmt.Sprintf("## %s\n", title) + body

		err := robot.SendMarkdown(title, body, nil, true)
		if err != nil {
			log.Error(err)
		}
	}
}
