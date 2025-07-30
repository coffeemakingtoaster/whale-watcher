package adapters_test

import (
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters"
)

func TestGitURLHttps(t *testing.T) {
	input := "https://github.com/coffeemakingtoaster/whale-watcher.git"
	expectedRepoId := "whale-watcher"
	expectedUser := "coffeemakingtoaster"
	actualUser, actualRepoId, err := adapters.ParseGitRepoURL("github.com", input)

	if err != nil {
		t.Fatal(err)
	}

	if actualRepoId != expectedRepoId {
		t.Errorf("Repo id mismatch: Expected %s Got %s", expectedRepoId, actualRepoId)
	}

	if actualUser != expectedUser {
		t.Errorf("User mismatch: Expected %s Got %s", actualUser, expectedUser)
	}

}

// Github does not support http (I think) this is for testing only
func TestGitURLHttp(t *testing.T) {
	input := "http://github.com/coffeemakingtoaster/whale-watcher.git"
	expectedRepoId := "whale-watcher"
	expectedUser := "coffeemakingtoaster"
	actualUser, actualRepoId, err := adapters.ParseGitRepoURL("github.com", input)

	if err != nil {
		t.Fatal(err)
	}

	if actualRepoId != expectedRepoId {
		t.Errorf("Repo id mismatch: Expected %s Got %s", expectedRepoId, actualRepoId)
	}

	if actualUser != expectedUser {
		t.Errorf("User mismatch: Expected %s Got %s", actualUser, expectedUser)
	}

}

// Github does not support http (I think) this is for testing only
func TestGitURLSSH(t *testing.T) {
	input := "git@github.com:coffeemakingtoaster/whale-watcher.git"
	expectedRepoId := "whale-watcher"
	expectedUser := "coffeemakingtoaster"
	actualUser, actualRepoId, err := adapters.ParseGitRepoURL("github.com", input)

	if err != nil {
		t.Fatal(err)
	}

	if actualRepoId != expectedRepoId {
		t.Errorf("Repo id mismatch: Expected %s Got %s", expectedRepoId, actualRepoId)
	}

	if actualUser != expectedUser {
		t.Errorf("User mismatch: Expected %s Got %s", actualUser, expectedUser)
	}

}
