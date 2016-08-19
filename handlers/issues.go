package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/LiberisLabs/github-webhooks/github"
)

type issueClient interface {
	GetIssue(issueURL string) (*github.Issue, error)
	UpdateIssue(issueURL string, update github.IssueUpdate) error
}

func handleIssues(storyRepo string, githubClient issueClient, logger *log.Logger, event github.IssueEvent) {
	if event.Action != "closed" || event.Repository.FullName == storyRepo {
		return
	}

	storyURL := findStory(storyRepo, event.Issue.Body)
	if storyURL == "" {
		logger.Println(event.Issue.URL, "could not find story:", event.Issue.Body)
		return
	}
	logger.Println(event.Issue.URL, "found connected story:", storyURL)

	story, err := githubClient.GetIssue(storyURL)
	if err != nil {
		logger.Println(event.Issue.URL, "error getting story:", err)
		return
	}

	newBody := tickReferencedIssue(story.Body, event.Repository.Owner.Login, event.Repository.Name, event.Issue.Number)

	if err := githubClient.UpdateIssue(story.URL, github.IssueUpdate{Body: newBody}); err != nil {
		logger.Println(event.Issue.URL, "error updating issue body:", err)
		return
	}
	logger.Println(event.Issue.URL, "updated story:", storyURL)
}

func findStory(storyRepo, body string) string {
	pattern := fmt.Sprintf(`https://github.com/%s/issues/(\d)|%s#(\d)`, storyRepo, storyRepo)
	expr := regexp.MustCompile(pattern)
	matches := expr.FindStringSubmatch(body)

	if matches == nil {
		return ""
	}

	storyNumber := matches[1]
	if matches[1] == "" {
		storyNumber = matches[2]
	}

	return fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", storyRepo, storyNumber)
}

func tickReferencedIssue(body, owner, repo string, number int) string {
	pattern := fmt.Sprintf(`- \[ \].+\(%s/%s#%d\)`, owner, repo, number)
	expr := regexp.MustCompile(pattern)

	return expr.ReplaceAllStringFunc(body, func(line string) string {
		return strings.Replace(line, "- [ ]", "- [x]", 1)
	})
}
