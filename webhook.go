package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/LiberisLabs/github-webhooks/github"
	"github.com/LiberisLabs/github-webhooks/handlers"
)

func randomCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)

	return base64.StdEncoding.EncodeToString(b)
}

type oauthHandler struct {
	clientID     string
	clientSecret string
	redirectURL  string
	onSuccess    func(accessToken string) http.Handler

	handler http.Handler
}

func (h *oauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.handler != nil {
		h.handler.ServeHTTP(w, r)
		return
	}

	query := url.Values{}
	query.Add("client_id", h.clientID)
	query.Add("redirect_uri", h.redirectURL+"/oauth/callback")

	switch r.URL.Path {
	case "/oauth":
		stateCookie, err := r.Cookie("state")
		if err != nil || stateCookie == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		redirectURL, _ := url.Parse("https://github.com/login/oauth/authorize")

		query.Add("state", stateCookie.Value)
		query.Add("scope", "repo")
		query.Add("allow_signup", "false")

		redirectURL.RawQuery = query.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)

	case "/oauth/callback":
		stateCookie, err := r.Cookie("state")
		if err != nil || stateCookie == nil || stateCookie.Value != r.FormValue("state") {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}

		tokenURL := "https://github.com/login/oauth/access_token"
		code := r.FormValue("code")

		query.Add("state", stateCookie.Value)
		query.Add("client_secret", h.clientSecret)
		query.Add("code", code)

		resp, err := http.PostForm(tokenURL, query)
		if err != nil {
			fmt.Fprint(w, "PostForm:", err)
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)
		form, _ := url.ParseQuery(string(body))

		h.handler = h.onSuccess(form.Get("access_token"))
		http.Redirect(w, r, "/", http.StatusFound)

	default:
		http.SetCookie(w, &http.Cookie{
			Name:     "state",
			Value:    randomCode(10),
			HttpOnly: true,
		})

		http.Redirect(w, r, "/oauth", http.StatusFound)
	}
}

func mustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Require environment variable %q to be set", key)
	}
	return value
}

func main() {
	port := os.Getenv("PORT")
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	storyRepo := mustGetenv("STORY_REPO")
	oauthClientID := mustGetenv("OAUTH_CLIENT_ID")
	oauthClientSecret := mustGetenv("OAUTH_CLIENT_SECRET")
	oauthRedirectURL := mustGetenv("OAUTH_REDIRECT_URL")

	if port == "" {
		port = "8080"
	}

	log.Println("Listening on :" + port)

	http.ListenAndServe(":"+port, &oauthHandler{
		clientID:     oauthClientID,
		clientSecret: oauthClientSecret,
		redirectURL:  oauthRedirectURL,
		onSuccess: func(accessToken string) http.Handler {
			gitHubClient := &github.Client{
				Client:    http.DefaultClient,
				Token:     accessToken,
				UserAgent: "gh-issues-flow golang",
			}

			return &handlers.Handler{
				GitHubClient: gitHubClient,
				Logger:       log.New(os.Stdout, "", log.LstdFlags),
				StoryRepo:    storyRepo,
				Secret:       []byte(secret),
			}
		},
	})
}
