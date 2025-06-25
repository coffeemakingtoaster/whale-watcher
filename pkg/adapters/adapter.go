package adapters

import "iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/adapters/github"

type PullRequestAdapter interface {
	CreatePullRequest() error
	UpdatePullRequest() error
}

func GetAdapterForRepository(repository string) PullRequestAdapter {
	return github.NewGithubPullRequestAdapter(repository)
}
