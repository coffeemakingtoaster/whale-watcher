package adapters

import (
	"errors"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters/github"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	violationTypes "github.com/coffeemakingtoaster/whale-watcher/pkg/validator/violations"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type PullRequestAdapter interface {
	CreatePullRequest(currentBranch, targetBranch, title, content string) error
	UpdatePullRequest(title, content string) error
	IsReady() bool
}

func GetAdapterForRepository(repository string) (PullRequestAdapter, error) {

	if config.ValidateGithub() == nil {
		if user, repo, err := ParseGitRepoURL("github.com", repository); err == nil {
			return github.NewGithubPullRequestAdapter(user, repo)
		}
	}

	if config.ValidateGitea() == nil {
		if user, repo, err := ParseGitRepoURL(viper.GetString("gitea.instanceurl"), repository); err == nil {
			return github.NewGithubPullRequestAdapter(user, repo)
		}

	}
	return &github.GithubPullRequestAdapter{}, errors.New("No configured vsc matched")
}

func CreatePRForFixes(violations violationTypes.Violations, updatedDockerfilePath string) error {
	// No fixes -> No Pr
	if len(violations.Violations) == 0 {
		log.Debug().Msg("No violations in current run, skipping PR creation")
		return nil
	}

	adapter, err := GetAdapterForRepository(viper.GetString("target.repository"))
	if err != nil {
		return err
	}

	if !adapter.IsReady() {
		log.Warn().Msg("Adapter was not ready, no git integration ran")
		// should this be an error?
		return nil
	}

	log.Debug().Msg("adapter is ready -> running git integration")

	newBranch, err := SyncFileToRepoIfDifferent(viper.GetString("target.repository"), viper.GetString("target.branch"), viper.GetString("target.dockerfilepath"), updatedDockerfilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new branch on remote for fixes PR")
		return err
	}
	if newBranch == "" {
		log.Info().Msg("No changes detected, remote and local file do not differ. No PR needed")
		return nil
	}

	err = adapter.CreatePullRequest(newBranch, viper.GetString("target.branch"), "Autofixes", violations.BuildDescriptionMarkdown())
	return err
}
