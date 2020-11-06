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

		err := handleHook(name)
		if err != nil {
			log.Error(err)
		}
	}
}

func handleHook(name string) error {

	var err error
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
		}

		for _, notification := range repository.Notifications {
			sendNotification(notification, buildOutput, err)
		}
	}()

	output, err := shellExec("git rev-parse --is-inside-work-tree", repository.Dir)
	if err != nil {
		panic(fmt.Sprintf("%s is not git repository: %s", repository.Dir, err.Error()))
	}
	if ! strings.Contains(output, "true") {
		panic(fmt.Sprintf("%s is not git repository: check %s", repository.Dir, output))
	}

	// pull code
	output, err = shellExec("git pull", repository.Dir)
	if err != nil {
		panic(fmt.Sprintf("git pull fail: %s", err.Error()))
	}
	if strings.Contains(output, "Already") {
		log.Info("git Already")
		return nil
	}

	// exec all shell
	for _, hook := range repository.Hooks {

		dir := repository.Dir
		if len(hook.Dir) > 0 {
			dir = hook.Dir
		}

		output, err := shellExec(hook.Shell, dir)
		if err != nil {
			panic(fmt.Sprintf("exec [%s] fail :%s", hook.Shell, err.Error()))
		}

		buildOutput = append(buildOutput, output)
		if len(hook.Assert) > 0 && !strings.Contains(output, hook.Assert) {
			if hook.AssertFailContinue {
				continue
			}

			panic(fmt.Sprintf("[output] %s \n[assert] %s\n", output, hook.Assert))
		}

		if len(hook.AssertNo) > 0 && strings.Contains(output, hook.AssertNo) {
			if hook.AssertFailContinue {
				continue
			}

			panic(fmt.Sprintf("[output] %s \n[assert_no] %s\n", output, hook.Assert))
		}
	}

	return nil
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

func sendNotification(n Notification, buildOutput []string, err error) {

	if n.Type == "dingtalk" {

		robot := dingrobot.NewRobot(n.WebHook)
		body := "- build steps \n"
		for _, o := range buildOutput {
			body += fmt.Sprintf("- %s\n", o)
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
