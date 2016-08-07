package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LiberisLabs/github-webhooks/github"
)

func TestPing(t *testing.T) {
	var buf bytes.Buffer

	s := httptest.NewServer(&Handler{
		GitHubClient: &github.Client{},
		Logger:       log.New(&buf, "", 0),
		StoryRepo:    "example/test",
	})
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, nil)
	req.Header.Add("X-GitHub-Event", "ping")

	http.DefaultClient.Do(req)

	if buf.String() != "pong\n" {
		t.Fatal("expected 'pong' to be logged")
	}
}
