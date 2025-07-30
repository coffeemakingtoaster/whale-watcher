package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
)

const prTitlePrefix = "[ww]"

func newClient(username, password string) (*gitea.Client, error) {
	cfg := config.GetConfig()
	return gitea.NewClient(cfg.Gitea.InstanceUrl, gitea.SetBasicAuth(username, password))
}

func checkForExistingPr(repoUser, repoID, username, password string) (int64, error) {
	client, err := newClient(username, password)
	if err != nil {
		return 0, err
	}
	prs, _, err := client.ListRepoPullRequests(repoUser, repoID, gitea.ListPullRequestsOptions{
		State: "open",
	})
	if err != nil {
		return 0, err
	}

	for _, pr := range prs {
		if pr.Head != nil && strings.HasPrefix(pr.Title, prTitlePrefix) {
			return pr.Index, nil
		}
	}

	return 0, nil // No existing PR found
}

func createPullRequest(repoUser, repoId, username, password, title, currentBranch, targetBranch, content string) (int64, error) {
	client, err := newClient(username, password)
	if err != nil {
		return 0, err
	}

	pr, _, err := client.CreatePullRequest(repoUser, repoId, gitea.CreatePullRequestOption{
		Head:  currentBranch,
		Base:  targetBranch,
		Title: fmt.Sprintf("%s - %s", prTitlePrefix, title),
		Body:  content,
	})
	if err != nil {
		return 0, err
	}

	return pr.Index, nil
}

func updatePullRequest(prId int64, repoUser, repoId, username, password, title, body string) error {
	client, err := newClient(username, password)
	if err != nil {
		return err
	}

	prUpdate := gitea.EditPullRequestOption{
		Title: title,
		Body:  body,
	}
	if title != "" {
		prUpdate.Title = fmt.Sprintf("%s - %s", prTitlePrefix, title)
	}
	if body != "" {
		prUpdate.Body = body
	}
	_, _, err = client.EditPullRequest(repoUser, repoId, prId, prUpdate)
	return err
}
