package adapters

import (
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/adapters/github"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/config"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/validator"
)

type PullRequestAdapter interface {
	CreatePullRequest(currentBranch, targetBranch, title, content string) error
	UpdatePullRequest(title, content string) error
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

	newBranch, err := SyncFileToRepoIfDifferent(cfg.Target.RepositoryURL, cfg.Target.Branch, cfg.Target.DockerfilePath, updatedDockerfilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new branch on remote for fixes PR")
		return err
	}
	if newBranch == "" {
		log.Info().Msg("No changes detected, remote and local file do not differ. No PR needed")
		return nil
	}
	adapter, err := GetAdapterForRepository(cfg.Target.RepositoryURL)
	if err != nil {
		return err
	}
	// TODO: This should likely actually detect the target branch
	err = adapter.CreatePullRequest(newBranch, "main", "fixes", "see violation text")
	return err
}
