# github-webhooks

Provides GitHub Webhook handlers, for:

- Marking a tasklist item as done in the story for a task when the task is
  closed.

```
$ go get github.com/LiberisLabs/github-webhooks
$ GITHUB_TOKEN='...' STORY_REPO='...' github-webhooks
...
```
