package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/LiberisLabs/github-webhooks/github"
)

type mockGitHubClient struct {
	issue *github.Issue

	getIssue struct {
		wasCalled bool
		issueURL  string
	}

	updateIssue struct {
		wasCalled bool
		issueURL  string
		update    github.IssueUpdate
	}
}

func (c *mockGitHubClient) GetIssue(issueURL string) (*github.Issue, error) {
	c.getIssue.wasCalled = true
	c.getIssue.issueURL = issueURL

	return c.issue, nil
}

func (c *mockGitHubClient) UpdateIssue(issueURL string, update github.IssueUpdate) error {
	c.updateIssue.wasCalled = true
	c.updateIssue.issueURL = issueURL
	c.updateIssue.update = update

	return nil
}

var logger = log.New(ioutil.Discard, "", 0)

func TestIssues(t *testing.T) {
	storyURL := "https://github.com/example/test/issues/5"
	storyApiURL := "https://api.github.com/repos/example/test/issues/5"

	body := `## Tasks

- [ ] This other task (example/something#1)
- [ ] This great task (example/something#5)
- [x] And a task I've already done (example/something#6)
`
	expectedNewBody := `## Tasks

- [ ] This other task (example/something#1)
- [x] This great task (example/something#5)
- [x] And a task I've already done (example/something#6)
`

	storyRepo := "example/test"
	githubClient := &mockGitHubClient{
		issue: &github.Issue{
			URL:    "http://example.com",
			ID:     2,
			Number: 3,
			Title:  "A cool issue",
			State:  "open",
			Body:   body,
		},
	}
	event := github.IssueEvent{
		Action: "closed",
		Issue: &github.Issue{
			Title:  "This great task",
			Number: 5,
			Body:   "## [Story](" + storyURL + ")\n\nWell this is a great task.",
		},
		Repository: &github.Repository{
			Name: "something",
			Owner: &github.Owner{
				Login: "example",
			},
		},
	}

	handleIssues(storyRepo, githubClient, logger, event)

	if githubClient.getIssue.issueURL != storyApiURL {
		t.Errorf("GetIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.getIssue.issueURL, storyApiURL)
	}

	if githubClient.updateIssue.issueURL != githubClient.issue.URL {
		t.Errorf("UpdateIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.updateIssue.issueURL, githubClient.issue.URL)
	}

	if githubClient.updateIssue.update.Body != expectedNewBody {
		t.Errorf("UpdateIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.updateIssue.update, github.IssueUpdate{Body: expectedNewBody})
	}
}

func TestIssuesWhenShortURLToStory(t *testing.T) {
	storyRepo := "example/test"
	storyNumber := 5
	storyApiURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", storyRepo, storyNumber)

	body := `## Tasks

- [ ] This other task (example/something#1)
- [ ] This great task (example/something#5)
- [x] And a task I've already done (example/something#6)
`
	expectedNewBody := `## Tasks

- [ ] This other task (example/something#1)
- [x] This great task (example/something#5)
- [x] And a task I've already done (example/something#6)
`

	githubClient := &mockGitHubClient{
		issue: &github.Issue{
			URL:    "http://example.com",
			ID:     2,
			Number: 3,
			Title:  "A cool issue",
			State:  "open",
			Body:   body,
		},
	}
	event := github.IssueEvent{
		Action: "closed",
		Issue: &github.Issue{
			Title:  "This great task",
			Number: 5,
			Body:   fmt.Sprintf("This does a fix for %s#%d ok?\n\nWell this is a great task.", storyRepo, storyNumber),
		},
		Repository: &github.Repository{
			Name: "something",
			Owner: &github.Owner{
				Login: "example",
			},
		},
	}

	handleIssues(storyRepo, githubClient, logger, event)

	if githubClient.getIssue.issueURL != storyApiURL {
		t.Errorf("GetIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.getIssue.issueURL, storyApiURL)
	}

	if githubClient.updateIssue.issueURL != githubClient.issue.URL {
		t.Errorf("UpdateIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.updateIssue.issueURL, githubClient.issue.URL)
	}

	if githubClient.updateIssue.update.Body != expectedNewBody {
		t.Errorf("UpdateIssue(...) was not called with the correct arguments\n     was\t%#v\nexpected\t%#v",
			githubClient.updateIssue.update, github.IssueUpdate{Body: expectedNewBody})
	}
}

func TestIssuesWhenEventNotClosed(t *testing.T) {
	storyRepo := "example/test"
	githubClient := &mockGitHubClient{}
	event := github.IssueEvent{
		Action: "opened",
	}

	handleIssues(storyRepo, githubClient, logger, event)

	if githubClient.getIssue.wasCalled || githubClient.updateIssue.wasCalled {
		t.Error("githubClient was called")
	}
}

func TestIssuesWhenIssueOnStoryRepo(t *testing.T) {
	storyRepo := "example/test"
	githubClient := &mockGitHubClient{}
	event := github.IssueEvent{
		Action: "closed",
		Repository: &github.Repository{
			FullName: storyRepo,
		},
	}

	handleIssues(storyRepo, githubClient, logger, event)

	if githubClient.getIssue.wasCalled || githubClient.updateIssue.wasCalled {
		t.Error("githubClient was called")
	}
}
