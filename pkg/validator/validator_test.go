package validator_test

import (
	"errors"
	"os"
	"testing"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/validator"
)

type MockRunner struct {
	callback func(bool) error
}

func (mr MockRunner) Run(runner.TemplateData, string) error { return mr.callback(false) }
func (mr MockRunner) RunFix(command string)                 { mr.callback(true) }
func (mr MockRunner) ToString() string                      { return "" }

func TestValidateFullValidRuleset(t *testing.T) {
	executionCount := 0
	validRunner := MockRunner{func(_ bool) error {
		executionCount++
		return nil
	},
	}
	input := rules.RuleSet{
		Rules: []*rules.Rule{
			{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "valid test 1",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "",
			},
			{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "valid test 2",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "",
			},
		},
	}
	actual := validator.ValidateRuleset(input, "", "", "")
	if actual.CheckedCount != 2 {
		t.Errorf("checkedcount mismatch: Expected %d Got %d", len(input.Rules), actual.CheckedCount)
	}
	if actual.ViolationCount != 0 {
		t.Errorf("violation mismatch: Expected 0 Got %d", actual.ViolationCount)
	}
	if executionCount != 2 {
		t.Errorf("Execution count mismatch: Expected 2 Got %d", executionCount)
	}
}

func TestValidateFullUnfixableFailingRuleset(t *testing.T) {
	runExecutionCount := 0
	fixExecutionCount := 0
	validRunner := MockRunner{func(is_test bool) error {
		if !is_test {
			runExecutionCount++
		} else {
			fixExecutionCount++
		}
		return errors.New("No")
	},
	}
	input := rules.RuleSet{
		Rules: []*rules.Rule{
			&rules.Rule{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "invalid test 1",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "",
			},
			&rules.Rule{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "invalid test 2",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "",
			},
		},
	}
	actual := validator.ValidateRuleset(input, "", "", "")
	if actual.CheckedCount != 2 {
		t.Errorf("checkedcount mismatch: Expected %d Got %d", len(input.Rules), actual.CheckedCount)
	}
	if actual.ViolationCount != 2 {
		t.Errorf("violation mismatch: Expected 0 Got %d", actual.ViolationCount)
	}
	if actual.FixableCount != 0 {
		t.Errorf("fixable mismatch: Expected 0 Got %d", actual.ViolationCount)
	}

	if runExecutionCount != 2 {
		t.Errorf("Execution count mismatch: Expected 2 Got %d", runExecutionCount)
	}

	if fixExecutionCount != 0 {
		t.Errorf("fix execution count mismatch: Expected 0 Got %d", fixExecutionCount)
	}
}

func TestValidateFullFixableFailingRuleset(t *testing.T) {
	runExecutionCount := 0
	fixExecutionCount := 0
	validRunner := MockRunner{func(is_test bool) error {
		if !is_test {
			runExecutionCount++
		} else {
			fixExecutionCount++
		}
		return errors.New("No")
	},
	}
	input := rules.RuleSet{
		Rules: []*rules.Rule{
			&rules.Rule{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "invalid test 1",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "xy",
			},
			&rules.Rule{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "invalid test 2",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "xy",
			},
		},
	}
	actual := validator.ValidateRuleset(input, "", "", "")
	if actual.CheckedCount != 2 {
		t.Errorf("checkedcount mismatch: Expected %d Got %d", len(input.Rules), actual.CheckedCount)
	}
	if actual.ViolationCount != 2 {
		t.Errorf("violation count mismatch: Expected 0 Got %d", actual.ViolationCount)
	}
	if actual.FixableCount != 2 {
		t.Errorf("fixable mismatch: Expected 0 Got %d", actual.ViolationCount)
	}

	if runExecutionCount != 2 {
		t.Errorf("Execution count mismatch: Expected 2 Got %d", runExecutionCount)
	}

	if fixExecutionCount != 2 {
		t.Errorf("fix execution count mismatch: Expected 2 Got %d", fixExecutionCount)
	}
}

func TestLimitTargetValidation(t *testing.T) {
	executionCount := 0
	validRunner := MockRunner{func(_ bool) error {
		executionCount++
		return nil
	},
	}
	input := rules.RuleSet{
		Rules: []*rules.Rule{
			{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "valid test 1",
				Target:         "fs",
				Runner:         validRunner,
				FixInstruction: "",
			},
			{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "valid test 2",
				Target:         "cmd",
				Runner:         validRunner,
				FixInstruction: "",
			},
			{
				Scope:          "output",
				Category:       "Negative",
				Instruction:    "abc",
				Description:    "def",
				Id:             "valid test 3",
				Target:         "os",
				Runner:         validRunner,
				FixInstruction: "",
			},
		},
	}
	os.Setenv("WHALE_WATCHER_TARGET_LIST", "cmd, fs")
	actual := validator.ValidateRuleset(input, "", "", "")
	if actual.CheckedCount != 2 {
		t.Errorf("checkedcount mismatch: Expected %d Got %d", len(input.Rules), actual.CheckedCount)
	}
	if actual.ViolationCount != 0 {
		t.Errorf("violation mismatch: Expected 0 Got %d", actual.ViolationCount)
	}
	if executionCount != 2 {
		t.Errorf("Execution count mismatch: Expected 2 Got %d", executionCount)
	}
}
