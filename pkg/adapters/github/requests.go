package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

const prTitlePrefix = "[ww]"

func checkForExistingPr(repoUser, repoID, pat string) (int, error) {
	ctx := context.Background()

	// Setup authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// List pull requests (you can paginate if needed)
	opts := &github.PullRequestListOptions{
		State:       "open", // you can also use "all" or "closed"
		ListOptions: github.ListOptions{PerPage: 100},
	}

	prs, _, err := client.PullRequests.List(ctx, repoUser, repoID, opts)
	if err != nil {
		return 0, err
	}

	// Check for PR with the given prefix
	for _, pr := range prs {
		if pr.Title != nil && strings.HasPrefix(*pr.Title, prTitlePrefix) {
			return *pr.Number, nil
		}
	}

	return 0, nil
}

func createPullRequest(repoUser, repoId, pat, title, currentBranch, targetBranch, content string) (int, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title:               github.String(fmt.Sprintf("%s - %s", prTitlePrefix, title)),
		Head:                github.String(currentBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(content),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, repoUser, repoId, newPR)
	if err != nil {
		return -1, err
	}

	return *pr.Number, nil
}

func updatePullRequest(prId int, repoUser, repoId, pat, title, body string) (*github.PullRequest, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	prUpdate := &github.PullRequest{}
	if title != "" {
		prUpdate.Title = github.String(fmt.Sprintf("%s - %s", prTitlePrefix, title))
	}
	if body != "" {
		prUpdate.Body = &body
	}

	pr, _, err := client.PullRequests.Edit(ctx, repoUser, repoId, int(prId), prUpdate)
	if err != nil {
		return nil, err
	}

	return pr, nil
}
