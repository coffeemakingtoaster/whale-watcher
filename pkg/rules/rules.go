package rules

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/fetcher"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func LoadRuleset(location string) (RuleSet, error) {
	ruleset, err := loadRuleSet(location)
	if err != nil {
		return RuleSet{}, err
	}
	if len(ruleset.Include) == 0 {
		return ruleset, nil
	}
	for i := range ruleset.Include {
		source := ruleset.Include[len(ruleset.Include)-1-i]
		weakSet, err := loadRuleSet(source)
		if err != nil {
			log.Error().Err(err).Str("source", source).Str("nestingset", ruleset.Name).Msg("Could not load included ruleset from source due to an erro")
			continue
		}
		ruleset.Swallow(weakSet)
	}
	return ruleset, nil
}

func loadRuleSet(location string) (RuleSet, error) {
	if strings.HasPrefix(location, "http://") {
		log.Debug().Msg("Provided ruleset location is a (unsafe) git repository!")
		if !util.IsUnsafeMode() {
			return RuleSet{}, errors.New("Could not load ruleset from unsafe repository (unsafe mode is disabled)")
		}
		return loadRuleSetFromRepository(location)
	}
	if strings.HasPrefix(location, "https://") {
		log.Debug().Msg("Provided ruleset location is a git repository!")
		return loadRuleSetFromRepository(location)
	}
	if strings.HasPrefix(location, "git@") {
		log.Debug().Msg("Provided ruleset location is a ssh git repository!")
		return loadRuleSetFromRepository(location)
	}
	log.Debug().Msg("Provided ruleset location is a filepath!")
	return loadRuleSetFromFile(location)
}

func loadRuleSetFromRepository(repositoryURL string) (RuleSet, error) {
	var ruleSet RuleSet
	url, internalPath, err := separateWWRepositoryUrlIntoGitURlAndInternalPath(repositoryURL)

	if err != nil {
		return ruleSet, err
	}

	if !strings.HasSuffix(url, ".git") {
		return ruleSet, fmt.Errorf("url (%s) has to end in .git", url)
	}

	data, err := fetcher.GetFileFromRepository(url, "main", internalPath)

	if err != nil {
		return ruleSet, err
	}

	return LoadRuleSetFromContent(data)
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

func separateWWRepositoryUrlIntoGitURlAndInternalPath(repositoryURl string) (string, string, error) {
	matcher := util.NewSliceSearch[rune]([]rune(".git!"))
	for i, c := range []rune(repositoryURl) {
		if matcher.Match(c) {
			return repositoryURl[:i], repositoryURl[i+1:], nil
		}
	}
	return "", "", errors.New(fmt.Sprintf(".git!<path> not contained in url %s", repositoryURl))
}
