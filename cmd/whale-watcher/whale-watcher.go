package main

import (
	"fmt"
	"os"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/validator"
)

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a ruleset location")
	}
	ruleSet, err := rules.LoadRuleSetFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d rules!\n", len(ruleSet.Rules))
	violations := validator.ValidateRuleset(ruleSet, "", "")
	fmt.Printf("Total: %d Violations: %d Fixable: %d\n", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
}
