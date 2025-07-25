package adapters

import (
	"github.com/rs/zerolog/log"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters/github"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
)

type PullRequestAdapter interface {
	CreatePullRequest(currentBranch, targetBranch, title, content string) error
	UpdatePullRequest(title, content string) error
	IsReady() bool
}

func GetAdapterForRepository(repository string) (PullRequestAdapter, error) {
	return github.NewGithubPullRequestAdapter(repository)
}

func CreatePRForFixes(violations validator.Violations, updatedDockerfilePath string) error {
	// No fixes -> No Pr
	if len(violations.Violations) == 0 {
		log.Debug().Msg("No violations in current run, skipping PR creation")
		return nil
	}
	cfg := config.GetConfig()

	adapter, err := GetAdapterForRepository(cfg.Target.RepositoryURL)
	if err != nil {
		return err
	}

	if !adapter.IsReady() {
		log.Warn().Msg("Adapter was not ready, no git integration ran")
		// should this be an error?
		return nil
	}

	log.Debug().Msg("adapter is ready -> running git integration")

	newBranch, err := SyncFileToRepoIfDifferent(cfg.Target.RepositoryURL, cfg.Target.Branch, cfg.Target.DockerfilePath, updatedDockerfilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new branch on remote for fixes PR")
		return err
	}
	if newBranch == "" {
		log.Info().Msg("No changes detected, remote and local file do not differ. No PR needed")
		return nil
	}

	err = adapter.CreatePullRequest(newBranch, cfg.Target.Branch, "Autofixes", violations.BuildDescriptionMarkdown())
	return err
}
