package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LiberisLabs/github-webhooks/github"
)

func TestPing(t *testing.T) {
	var buf bytes.Buffer
	expectedMessage := "pong\n"

	s := httptest.NewServer(&Handler{
		GitHubClient: &github.Client{},
		Logger:       log.New(&buf, "", 0),
		StoryRepo:    "example/test",
	})
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, nil)
	req.Header.Add("X-GitHub-Event", "ping")

	http.DefaultClient.Do(req)

	if buf.String() != expectedMessage {
		t.Fatalf("expected %q to be logged, but got %q", expectedMessage, buf.String())
	}
}

func TestPingWithSignature(t *testing.T) {
	var buf bytes.Buffer
	expectedMessage := "pong\n"

	s := httptest.NewServer(&Handler{
		GitHubClient: &github.Client{},
		Logger:       log.New(&buf, "", 0),
		StoryRepo:    "example/test",
		Secret:       []byte("this is a secret"),
	})
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, strings.NewReader("this is the body"))
	req.Header.Add("X-GitHub-Event", "ping")
	req.Header.Add("X-Hub-Signature", "sha1=cd85d074acb477fdd414a5009290a9cec17cc8b1")

	http.DefaultClient.Do(req)

	if buf.String() != expectedMessage {
		t.Fatalf("expected %q to be logged, but got %q", expectedMessage, buf.String())
	}
}

func TestPingWithInvalidSignature(t *testing.T) {
	var buf bytes.Buffer
	expectedMessage := "pong\nping: invalid signature\n"

	s := httptest.NewServer(&Handler{
		GitHubClient: &github.Client{},
		Logger:       log.New(&buf, "", 0),
		StoryRepo:    "example/test",
		Secret:       []byte("this is a secret"),
	})
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, strings.NewReader("this is the body"))
	req.Header.Add("X-GitHub-Event", "ping")
	req.Header.Add("X-Hub-Signature", "sha1=cd85d074acb477fdd414a5009290a9cec17cc8b2")

	http.DefaultClient.Do(req)

	if buf.String() != expectedMessage {
		t.Fatalf("expected %q to be logged, but got %q", expectedMessage, buf.String())
	}
}
