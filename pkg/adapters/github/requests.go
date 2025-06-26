package github

import (
	"context"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

func createPullRequest(repoUser, repoId, pat, title, currentBranch, targetBranch, content string) (int64, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(currentBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(content),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, repoUser, repoId, newPR)
	if err != nil {
		return -1, err
	}

	return *pr.ID, nil
}
