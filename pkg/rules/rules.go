package rules

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func LoadRuleSetFromRepository(repositoryURL, path string) (RuleSet, error) {
	// TODO: implement me
	return RuleSet{}, nil
}

func LoadRuleSetFromFile(path string) (RuleSet, error) {
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
		if !strings.Contains(v.Instruction, "assert") {
			log.Warn().Str("Instruction", v.Instruction).Msg("Instruction does not contain an assert. This rule therefore will never be checked properly")
		}
	}
	return ruleSet, nil
}
