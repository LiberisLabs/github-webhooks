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
	"github.com/gorilla/securecookie"
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
	secureCookie *securecookie.SecureCookie
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
		expectedState, err := getSecureCookie(h.secureCookie, r, "state")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		redirectURL, _ := url.Parse("https://github.com/login/oauth/authorize")

		query.Add("state", expectedState)
		query.Add("scope", "repo")
		query.Add("allow_signup", "false")

		redirectURL.RawQuery = query.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)

	case "/oauth/callback":
		expectedState, err := getSecureCookie(h.secureCookie, r, "state")
		if err != nil || expectedState != r.FormValue("state") {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}

		tokenURL := "https://github.com/login/oauth/access_token"
		code := r.FormValue("code")

		query.Add("state", expectedState)
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
		if err := setSecureCookie(h.secureCookie, w, "state", randomCode(10)); err == nil {
			http.Redirect(w, r, "/oauth", http.StatusFound)
		}
	}
}

func getSecureCookie(s *securecookie.SecureCookie, r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	var value string
	err = s.Decode(name, cookie.Value, &value)
	return value, err
}

func setSecureCookie(s *securecookie.SecureCookie, w http.ResponseWriter, name, value string) error {
	encoded, err := s.Encode(name, value)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    encoded,
		HttpOnly: true,
	})

	return nil
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

	hashKey := []byte(mustGetenv("COOKIE_HASH_KEY"))
	blockKey := []byte(os.Getenv("COOKIE_BLOCK_KEY"))
	if len(blockKey) == 0 {
		blockKey = nil
	}

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
		secureCookie: securecookie.New(hashKey, blockKey),
		onSuccess: func(accessToken string) http.Handler {
			gitHubClient := &github.Client{
				Client:    http.DefaultClient,
				Token:     accessToken,
				UserAgent: "github.com/LiberisLabs/github-webhooks golang",
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
