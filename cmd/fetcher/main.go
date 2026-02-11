package main

import (
	"github-weekly-log/internal/github"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	GITHUB_USER := os.Getenv("GITHUB_USER")

	client := github.NewClient(GITHUB_TOKEN)

	evnets, err := client.FetchWeeklyEvents(GITHUB_USER)
	if err != nil {
		panic(err)
	}

	_ = evnets

}
