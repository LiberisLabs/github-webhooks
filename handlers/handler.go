package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/LiberisLabs/github-webhooks/github"
)

type Handler struct {
	GitHubClient *github.Client
	Logger       *log.Logger
	StoryRepo    string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event")

	body := json.NewDecoder(r.Body)
	defer r.Body.Close()

	switch event {
	case "ping":
		h.Logger.Println("pong")

	case "issues":
		var v github.IssueEvent
		body.Decode(&v)
		go handleIssues(h.StoryRepo, h.GitHubClient, h.Logger, v)
	}

	w.WriteHeader(http.StatusOK)
}
