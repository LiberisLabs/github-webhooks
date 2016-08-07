package github

import (
	"bytes"
	"encoding/json"
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
	req.Header.Add("Authorization", "Bearer "+c.Token)
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
