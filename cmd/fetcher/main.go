package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v60/github"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	GITHUB_USER := os.Getenv("GITHUB_USER")

	client := github.NewClient(nil).WithAuthToken(GITHUB_TOKEN)

	opt := &github.RepositoryListByUserOptions{
		Type: "owner",
	}
	repos, _, err := client.Repositories.ListByUser(context.Background(), GITHUB_USER, opt)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(repos))
	for i := range repos {
		fmt.Println(*repos[i].Name)
	}
	_ = repos
}
