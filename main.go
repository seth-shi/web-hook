package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
)


func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

}

func main() {

	r, err := git.PlainOpen(os.Getenv("CODE_PATH"))
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(r)


	router := gin.Default()
	router.GET("/ping", ping)

	err = router.Run(fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalln(err)
	}
}

func ping(c *gin.Context)  {

	c.JSON(200, gin.H{
		"message": "pong",
	})
}
