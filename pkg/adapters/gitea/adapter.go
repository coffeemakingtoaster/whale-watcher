package gitea

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type GiteaPullRequestAdapter struct {
	repoUser         string
	repoId           string
	whaleWatcherUser string
	authUsername     string
	authPassword     string
}

func (gtpra *GiteaPullRequestAdapter) IsReady() bool {
	existingPrId, err := checkForExistingPr(gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword)
	log.Debug().Err(err).Int("prid", int(existingPrId)).Msg("Result of checking for existing PRs")
	// No existing PR and no error -> go ahead
	return existingPrId == 0 && err == nil
}

func (gtpra *GiteaPullRequestAdapter) CreatePullRequest(currentBranch, targetBranch, title, content string) error {
	existingPrId, err := checkForExistingPr(gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword)
	if existingPrId != 0 && err == nil {
		err := updatePullRequest(existingPrId, gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword, title, content)
		return err
	} else {
		_, err := createPullRequest(gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword, title, currentBranch, targetBranch, content)
		return err
	}
}

func (gtpra *GiteaPullRequestAdapter) UpdatePullRequest(title, content string) error {
	existingPrId, err := checkForExistingPr(gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword)
	if err != nil {
		return err
	}
	err = updatePullRequest(existingPrId, gtpra.repoUser, gtpra.repoId, gtpra.authUsername, gtpra.authPassword, title, content)
	return err
}

func NewGiteaPullRequestAdapter(repoUser, repoId string) (*GiteaPullRequestAdapter, error) {

	return &GiteaPullRequestAdapter{
		repoUser:         repoUser,
		repoId:           repoId,
		whaleWatcherUser: viper.GetString("gitea.username"),
		authUsername:     viper.GetString("gitea.username"),
		authPassword:     viper.GetString("gitea.password"),
	}, nil
}
