package github

import (
	"errors"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/rs/zerolog/log"
)

type GithubPullRequestAdapter struct {
	pat              string
	repoUser         string
	repoId           string
	whaleWatcherUser string
}

func (ghpra *GithubPullRequestAdapter) IsReady() bool {
	existingPrId, err := checkForExistingPr(ghpra.repoUser, ghpra.repoId, ghpra.pat)
	log.Debug().Err(err).Int("prid", int(existingPrId)).Send()
	// No existing PR and no error -> go ahead
	return existingPrId == 0 && err == nil
}

func (ghpra *GithubPullRequestAdapter) CreatePullRequest(currentBranch, targetBranch, title, content string) error {
	existingPrId, err := checkForExistingPr(ghpra.repoUser, ghpra.repoId, ghpra.pat)
	if existingPrId != 0 && err == nil {
		_, err := updatePullRequest(existingPrId, ghpra.repoUser, ghpra.repoId, ghpra.pat, title, content)
		return err
	} else {
		_, err := createPullRequest(ghpra.repoUser, ghpra.repoId, ghpra.pat, title, currentBranch, targetBranch, content)
		return err
	}
}

func (ghpra *GithubPullRequestAdapter) UpdatePullRequest(title, content string) error {
	existingPrId, err := checkForExistingPr(ghpra.repoUser, ghpra.repoId, ghpra.pat)
	if err != nil {
		return err
	}
	_, err = updatePullRequest(existingPrId, ghpra.repoUser, ghpra.repoId, ghpra.pat, title, content)
	return err
}

func NewGithubPullRequestAdapter(repoUser, repoId string) (*GithubPullRequestAdapter, error) {
	conf := config.GetConfig()

	if len(conf.Github.PAT) == 0 || len(conf.Github.Username) == 0 {
		return nil, errors.New("Some required github config fields not set!")
	}

	return &GithubPullRequestAdapter{
		repoUser:         repoUser,
		repoId:           repoId,
		pat:              conf.Github.PAT,
		whaleWatcherUser: conf.Github.Username,
	}, nil
}
