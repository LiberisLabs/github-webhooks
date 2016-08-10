package github

type IssueEvent struct {
	Action     string      `json:"action"`
	Issue      *Issue      `json:"issue"`
	Repository *Repository `json:"repository"`
	Label      *Label      `json:"label"`
}

type Issue struct {
	URL    string   `json:"url"`
	ID     int      `json:"id"`
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Labels []*Label `json:"labels"`
	State  string   `json:"state"`
	Body   string   `json:"body"`
}

type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    *Owner `json:"owner"`
}

type Owner struct {
	Login string `json:"login"`
}

type Label struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type IssueUpdate struct {
	Body string `json:"body"`
}
