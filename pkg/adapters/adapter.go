package adapters

import "iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/adapters/github"

type PullRequestAdapter interface {
	CreatePullRequest(currentBranch, targetBranch, title, content string) error
	UpdatePullRequest(title, content string) error
}

func GetAdapterForRepository(repository string) (PullRequestAdapter, error) {
	return github.NewGithubPullRequestAdapter(repository)
}
