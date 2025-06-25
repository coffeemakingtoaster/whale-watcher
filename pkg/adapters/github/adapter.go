package github

type GithubPullRequestAdapter struct {
	pat        string
	repository string
}

func (ghpra *GithubPullRequestAdapter) CreatePullRequest() error { return nil }

func (ghpra *GithubPullRequestAdapter) UpdatePullRequest() error { return nil }

func NewGithubPullRequestAdapter(repository string) *GithubPullRequestAdapter {
	return &GithubPullRequestAdapter{}
}
