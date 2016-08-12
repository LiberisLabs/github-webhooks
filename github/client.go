package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const acceptMediaType = "application/vnd.github.v3+json"

type Client struct {
	Client    *http.Client
	Token     string
	UserAgent string
}

func (c *Client) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", acceptMediaType)
	req.Header.Add("Authorization", "token "+c.Token)
	req.Header.Add("User-Agent", c.UserAgent)

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return resp, err
		}
	}

	return resp, err
}

func (c *Client) GetIssue(issueURL string) (*Issue, error) {
	req, err := c.newRequest("GET", issueURL, nil)
	if err != nil {
		return nil, err
	}

	var v Issue
	if _, err := c.do(req, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) UpdateIssue(issueURL string, update IssueUpdate) error {
	req, err := c.newRequest("PATCH", issueURL, &update)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

func (c *Client) ListRepositories(organization string) ([]*Repository, error) {
	var repos []*Repository

	req, err := c.newRequest("GET", fmt.Sprintf("https://api.github.com/orgs/%s/repos", organization), nil)
	if err != nil {
		return repos, err
	}

	_, err = c.do(req, &repos)
	return repos, err
}

func (c *Client) GetHooks(hooksURL string) ([]*Hook, error) {
	var hooks []*Hook

	req, err := c.newRequest("GET", hooksURL, nil)
	if err != nil {
		return hooks, err
	}

	_, err = c.do(req, &hooks)
	return hooks, err
}

func (c *Client) InstallWebhook(webhookURL, organization, secret string) error {
	repos, err := c.ListRepositories(organization)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		hooks, _ := c.GetHooks(repo.HooksURL)
		alreadyInstalled := false

		for _, hook := range hooks {
			if hook.Config.URL == webhookURL {
				alreadyInstalled = true
				break
			}
		}

		if !alreadyInstalled {
			c.CreateHook(repo.HooksURL, Hook{
				Active: true,
				Name:   "web",
				Events: []string{"issues"},
				Config: HookConfig{
					URL:         webhookURL,
					ContentType: "json",
					Secret:      secret,
				},
			})
		}
	}

	return nil
}

func (c *Client) CreateHook(hooksURL string, hook Hook) error {
	req, err := c.newRequest("POST", hooksURL, &hook)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)

	return err
}
