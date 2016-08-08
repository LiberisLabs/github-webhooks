package main

import (
	"log"
	"net/http"
	"os"

	"github.com/LiberisLabs/github-webhooks/github"
	"github.com/LiberisLabs/github-webhooks/handlers"
)

func main() {
	port := os.Getenv("PORT")
	token := os.Getenv("GITHUB_TOKEN")
	secret := os.Getenv("GITHUB_SECRET")
	storyRepo := os.Getenv("STORY_REPO")

	if port == "" {
		port = "8080"
	}

	if token == "" || storyRepo == "" {
		log.Fatal("Must provide GITHUB_TOKEN and STORY_REPO environment variables")
	}

	gitHubClient := &github.Client{
		Client:    http.DefaultClient,
		Token:     token,
		UserAgent: "gh-issues-flow golang",
	}

	log.Println("Listening on :" + port)
	http.ListenAndServe(":"+port, http.Handler(&handlers.Handler{
		GitHubClient: gitHubClient,
		Logger:       log.New(os.Stdout, "", log.LstdFlags),
		StoryRepo:    storyRepo,
		Secret:       []byte(secret),
	}))
}
