package rules

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"github.com/coffeemakingtoaster/whale-watcher/internal/environment"
)

func LoadRuleset(location string) (RuleSet, error) {
	if strings.HasPrefix(location, "http://") {
		log.Debug().Msg("Provided ruleset location is a (unsafe) git repository!")
		if !environment.IsUnsafeMode() {
			return RuleSet{}, errors.New("Could not load ruleset from unsafe repository (unsafe mode is disabled)")
		}
		return loadRuleSetFromRepository(location)
	}
	if strings.HasPrefix(location, "https://") {
		log.Debug().Msg("Provided ruleset location is a git repository!")
		return loadRuleSetFromRepository(location)
	}
	log.Debug().Msg("Provided ruleset location is a filepath!")
	return loadRuleSetFromFile(location)
}

func loadRuleSetFromRepository(repositoryURL string) (RuleSet, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	repository, err := git.Clone(storer, fs, &git.CloneOptions{URL: repositoryURL})

	if err != nil {
		return RuleSet{}, err
	}

	head, _ := repository.Head()
	log.Info().Str("Hash", head.Hash().String()).Msgf("Using remote state for repository '%s'", repositoryURL)

	var ruleSet RuleSet

	files, err := fs.ReadDir(".")

	for _, entry := range files {
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			log.Debug().Str("filename", entry.Name()).Msg("Skipped remote file")
			continue
		}

		log.Debug().Str("filename", entry.Name()).Msg("Parsing remote file")
		fileHandle, err := fs.Open(entry.Name())
		if err != nil {
			return RuleSet{}, err
		}

		var data bytes.Buffer

		_, err = io.Copy(&data, fileHandle)

		if err != nil {
			return RuleSet{}, err
		}

		log.Debug().Str("filename", entry.Name()).Int("filesize", data.Len()).Msg("Parsed file of size")

		ruleSet, err = LoadRuleSetFromContent(data.Bytes())
		if err != nil {
			return RuleSet{}, err
		}
		return ruleSet, nil
	}

	return ruleSet, nil
}

func loadRuleSetFromFile(path string) (RuleSet, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return RuleSet{}, err
	}
	return LoadRuleSetFromContent(file)
}

func LoadRuleSetFromContent(data []byte) (RuleSet, error) {
	var ruleSet RuleSet

	err := yaml.Unmarshal(data, &ruleSet)
	if err != nil {
		return RuleSet{}, err
	}
	for _, v := range ruleSet.Rules {
		err := v.AddRunner()
		if err != nil {
			return RuleSet{}, err
		}
		err = v.Verify()
		if err != nil {
			return RuleSet{}, err
		}
		if !strings.Contains(v.Instruction, "assert") {
			log.Warn().Str("Instruction", v.Instruction).Msg("Instruction does not contain an assert. This rule therefore will never be checked properly")
		}
	}
	return ruleSet, nil
}
