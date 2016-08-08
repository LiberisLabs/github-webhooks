package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/LiberisLabs/github-webhooks/github"
	"github.com/LiberisLabs/github-webhooks/handlers"
)

var state = "rgerehureghurgehurge"

type oauthHandler struct {
	clientID     string
	clientSecret string
}

func (h *oauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := url.Values{}
	query.Add("client_id", h.clientID)
	query.Add("state", state)

	callbackURL, _ := r.URL.Parse("/oauth/callback")

	switch r.URL.Path {
	case "/":
		redirectURL, _ := url.Parse("https://github.com/login/oauth/authorize")

		query.Add("redirect_uri", callbackURL.String())
		query.Add("scope", "repo")
		query.Add("allow_signup", "false")

		redirectURL.RawQuery = query.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)

	case "/oauth/callback":
		code := r.FormValue("code")

		if r.FormValue("state") != state {
			return
		}

		tokenURL := "https://github.com/login/oauth/access_token"

		query.Add("client_secret", h.clientSecret)
		query.Add("code", code)
		query.Add("redirect_uri", callbackURL.String())

		resp, err := http.Post(tokenURL, "application/x-www-urlencoded-form", query.Encode())
		if err != nil {
			return
		}

		accessToken := resp.PostFormValue("access_token")
	}
}

func main() {
	port := os.Getenv("PORT")

	token := os.Getenv("GITHUB_TOKEN")
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	storyRepo := os.Getenv("STORY_REPO")

	oauthClientID := os.Getenv("OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	getAuth := os.Getenv("GET_AUTH")

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

	if getAuth != "" {
		http.ListenAndServe(":"+port, &oauthHandler{
			clientID:     oauthClientID,
			clientSecret: oauthClientSecret,
		})
	} else {
		http.ListenAndServe(":"+port, &handlers.Handler{
			GitHubClient: gitHubClient,
			Logger:       log.New(os.Stdout, "", log.LstdFlags),
			StoryRepo:    storyRepo,
			Secret:       []byte(secret),
		})
	}
}
