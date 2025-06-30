package github

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/config"
)

type GithubPullRequestAdapter struct {
	pat              string
	repoUser         string
	repoId           string
	whaleWatcherUser string
}

func (ghpra *GithubPullRequestAdapter) CreatePullRequest(currentBranch, targetBranch, title, content string) error {
	_, err := createPullRequest(ghpra.repoUser, ghpra.repoId, ghpra.pat, title, currentBranch, targetBranch, content)
	return err
}

func (ghpra *GithubPullRequestAdapter) UpdatePullRequest(title, content string) error { return nil }

func NewGithubPullRequestAdapter(repositoryURL string) (*GithubPullRequestAdapter, error) {
	repoUser, repoId, err := parseGitHubRepo(repositoryURL)
	if err != nil {
		return nil, err
	}
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

func parseGitHubRepo(repoURL string) (user, repo string, err error) {
	repoURL = strings.TrimSuffix(repoURL, ".git")

	if strings.HasPrefix(repoURL, "git@") {
		re := regexp.MustCompile(`git@github\.com:([^/]+)/(.+)`)
		matches := re.FindStringSubmatch(repoURL)
		if len(matches) == 3 {
			return matches[1], matches[2], nil
		}
		return "", "", errors.New("invalid SSH GitHub URL format")
	}

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}
	if parsedURL.Host != "github.com" {
		return "", "", errors.New("not a github.com URL")
	}
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid GitHub URL path")
	}
	return parts[0], parts[1], nil
}
