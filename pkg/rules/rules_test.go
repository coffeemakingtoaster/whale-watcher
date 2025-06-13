package rules_test

import (
	"reflect"
	"testing"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
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
				Scope:       "output",
				Category:    "negative",
				Instruction: "assert(True == True)\n",
				Description: "Perform a check",
				Id:          "test id",
				Target:      "command",
			},
			{
				Scope:       "buildtime",
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
	for index := range expected.Rules {
		actual.Rules[index].Runner = nil
		if !reflect.DeepEqual(expected.Rules[index], actual.Rules[index]) {
			t.Errorf("Rule content mismatch:\n Expected\n %v Got\n %v", expected.Rules[index], actual.Rules[index])
		}
	}
}

func TestVerifyInvalidRuleset(t *testing.T) {
	expected := map[string]rules.Rule{
		"No id set for rule": {
			Scope:       "buildtime",
			Category:    "positive",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Target:      "fs",
		},
		"Target: Invalid value invalid (Allowed: [\"command\" \"os\" \"fs\"])": {
			Scope:       "buildtime",
			Category:    "positive",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Id:          "test id2",
			Target:      "invalid",
		},
		"Category: Invalid value maybe (Allowed: [\"negative\" \"positive\"])": {
			Scope:       "buildtime",
			Category:    "maybe",
			Instruction: "assert(True == False)",
			Description: "Perform a check",
			Id:          "test id2",
			Target:      "fs",
		},
		"Scope: Invalid value never (Allowed: [\"output\" \"buildtime\"])": {
			Scope:       "never",
			Category:    "negative",
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
