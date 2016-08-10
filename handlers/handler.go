package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/LiberisLabs/github-webhooks/github"
)

type Handler struct {
	GitHubClient *github.Client
	Logger       *log.Logger
	StoryRepo    string
	Secret       []byte
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprint(w, "Ready...")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	event := r.Header.Get("X-GitHub-Event")

	var body io.Reader = r.Body
	defer r.Body.Close()

	var mac hash.Hash
	var sig []byte
	if len(h.Secret) > 0 {
		sig = []byte(r.Header.Get("X-Hub-Signature"))
		mac = hmac.New(sha1.New, h.Secret)
		body = io.TeeReader(r.Body, mac)
	}

	switch event {
	case "ping":
		h.Logger.Println("pong")

		if mac != nil {
			io.Copy(ioutil.Discard, body)

			if !verifySignature(mac, sig) {
				h.Logger.Println("ping: invalid signature")
				return
			}
		}

	case "issues":
		var v github.IssueEvent
		json.NewDecoder(body).Decode(&v)

		if mac != nil {
			if !verifySignature(mac, sig) {
				h.Logger.Println("issues: invalid signature")
				return
			}
		}

		go handleIssues(h.StoryRepo, h.GitHubClient, h.Logger, v)
	}

	w.WriteHeader(http.StatusOK)
}

func verifySignature(mac hash.Hash, expectedSig []byte) bool {
	sum := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal(expectedSig, []byte("sha1="+sum))
}
