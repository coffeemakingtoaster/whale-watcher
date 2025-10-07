package rules_test

import (
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/spf13/viper"
)

var noAssertRuleset = `
name: no assert ruleset
rules:
  - scope: output
    category: Negative
    instruction: |
        True == True
    description: Do nothing really
    id: no assert 
    target: command
`

var validRuleset = `
name:  test ruleset
rules:
 - scope: output
   category: Negative
   instruction: |
       assert(True == True)
   description: Perform a check
   id: test id
   target: command
 - scope: buildtime
   category: Positive
   instruction: |
       assert(True == False)
   description: Perform a check
   id: test id2
   target: fs
`

func TestLoadRulesetFromContent(t *testing.T) {
	expected := rules.RuleSet{
		Name: "test ruleset",
		Rules: []*rules.Rule{
			{
				Category:    "negative",
				Instruction: "assert(True == True)\n",
				Description: "Perform a check",
				Id:          "test id",
				Target:      "command",
			},
			{
				Category:    "positive",
				Instruction: "assert(True == False)\n",
				Description: "Perform a check",
				Id:          "test id2",
				Target:      "fs",
			},
		},
	}
	actual, err := rules.LoadRuleSetFromContent([]byte(validRuleset))
	if err != nil {
		t.Errorf("Error mismatch: Expected nil Got '%s'", err.Error())
	}
	if expected.Name != actual.Name {
		t.Errorf("Ruleset name mismatch: Expected %s Got %s", expected.Name, actual.Name)
	}
	if actual.GetHighestTarget() != "fs" {
		t.Errorf("Highest target mismatch: Expected fs Got %s", actual.Name)
	}
}

func TestHighestLevelIfConfigDisallow(t *testing.T) {
	expected := rules.RuleSet{
		Name: "test ruleset",
		Rules: []*rules.Rule{
			{
				Category:    "negative",
				Instruction: "assert(True == True)\n",
				Description: "Perform a check",
				Id:          "test id",
				Target:      "command",
			},
			{
				Category:    "positive",
				Instruction: "assert(True == False)\n",
				Description: "Perform a check",
				Id:          "test id2",
				Target:      "fs",
			},
		},
	}
	viper.Set("target_list", "command")
	defer viper.Reset()
	actual, err := rules.LoadRuleSetFromContent([]byte(validRuleset))
	if err != nil {
		t.Errorf("Error mismatch: Expected nil Got '%s'", err.Error())
	}
	if expected.Name != actual.Name {
		t.Errorf("Ruleset name mismatch: Expected %s Got %s", expected.Name, actual.Name)
	}
	if actual.GetHighestTarget() != "command" {
		t.Errorf("Highest target mismatch: Expected command Got %s", actual.Name)
	}
}

func TestVerifyInvalidRuleset(t *testing.T) {
	expected := map[string]rules.Rule{
		"No id set for rule": {
			Category:    "positive",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Target:      "fs",
		},
		"Target: Invalid value invalid (Allowed: [\"command\" \"os\" \"fs\"])": {
			Category:    "positive",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Id:          "test id2",
			Target:      "invalid",
		},
		"Category: Invalid value maybe (Allowed: [\"negative\" \"positive\"])": {
			Category:    "maybe",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Id:          "test id2",
			Target:      "fs",
		},
	}

	for errorMessage, rule := range expected {
		err := rule.Verify()
		if err.Error() != errorMessage {
			t.Errorf("Error mismatch: Expected %s Got %s", errorMessage, err.Error())
		}
	}
}

func TestLoadRulesetFromInvalidGitRepositoryUrl(t *testing.T) {
	_, err := rules.LoadRuleset("git@github.com:coffeemakingtoaster/whale-watcher.g!abc.yaml")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLoadRulesetFromGitRepositoryUrlWithoutPath(t *testing.T) {
	_, err := rules.LoadRuleset("git@github.com:coffeemakingtoaster/whale-watcher.git")
	if err == nil {
		t.Error("Expected error, got nil")
	}

}
