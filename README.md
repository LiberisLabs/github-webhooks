# github-webhooks

Provides GitHub Webhook handlers, for:

- Marking a tasklist item as done in the story for a task when the task is
  closed.

```
$ go get github.com/LiberisLabs/github-webhooks
$ github-webhooks
...
```

Requires the following environment variables:

- `STORY_REPO`: the full name of the repository used for tracking stories,
  e.g. `myorg/stories`
- `OAUTH_CLIENT_ID`
- `OAUTH_CLIENT_SECRET`
- `OAUTH_REDIRECT_URL`: the base URL this is hosted at without trailing slash,
  e.g. `https://example.com`

And can also use the following optional variables:

- `PORT`
- `GITHUB_WEBHOOK_SECRET`: the shared secret for verifying authenticity of
  webhooks
