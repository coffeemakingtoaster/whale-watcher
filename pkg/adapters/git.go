package adapters

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog/log"
)

func SyncFileToRepoIfDifferent(repoURL, branch, repoFilePath, hostFilePath string) (string, error) {
	// Clone the repo to a temporary directory
	fs := memfs.New()
	storer := memory.NewStorage()

	repository, err := git.Clone(
		storer,
		fs,
		&git.CloneOptions{URL: repoURL, NoCheckout: true},
	)

	if err != nil {
		return "", err
	}
	w, err := repository.Worktree()
	if err != nil {
		return "", err
	}
	dockerfileDirectory := filepath.Dir(repoFilePath)

	branchReference := plumbing.NewBranchReferenceName(branch)

	err = w.Checkout(&git.CheckoutOptions{SparseCheckoutDirectories: []string{dockerfileDirectory}, Branch: plumbing.ReferenceName(branchReference)})
	if err != nil {
		return "", err
	}

	f, err := fs.Open(repoFilePath)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	hostfileData, err := getFileData(hostFilePath)

	if err != nil {
		return "", err
	}

	filesEqual := bytes.Equal(hostfileData, data)

	if filesEqual {
		log.Debug().Msg("No changes to dockerfile detected. No PR needed")
		return "", nil
	}

	// Files differ — create a new branch
	newBranch := fmt.Sprintf("update-%d", time.Now().Unix())
	newRef := plumbing.NewBranchReferenceName(newBranch)

	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: newRef,
	})
	if err != nil {
		return "", fmt.Errorf("creating new branch: %w", err)
	}

	dockerFile, err := fs.OpenFile(repoFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}

	dockerFile.Write(hostfileData)

	dockerFile.Close()

	// Stage and commit the change
	_, err = w.Add(repoFilePath)
	if err != nil {
		return "", fmt.Errorf("adding file to git: %w", err)
	}

	_, err = w.Commit("Update file from host system", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Auto Commit Bot",
			Email: "bot@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("committing changes: %w", err)
	}

	// Push the new branch
	err = repository.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(newRef + ":" + newRef),
		},
	})
	if err != nil {
		return "", fmt.Errorf("pushing branch: %w", err)
	}

	fmt.Printf("Pushed updated file to branch '%s'\n", newBranch)
	return newBranch, nil
}

func getFileData(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
