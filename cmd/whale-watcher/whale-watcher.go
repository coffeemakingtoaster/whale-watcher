package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/validator"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if len(os.Args) < 2 {
		panic("Please provide a ruleset location")
	}
	ruleSet, err := rules.LoadRuleSetFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("Loaded %d rules!", len(ruleSet.Rules))
	violations := validator.ValidateRuleset(ruleSet, "", "")
	log.Info().Msgf("Total: %d Violations: %d Fixable: %d", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
	for _, violation := range violations.Violations {
		log.Warn().Str("ruleId", violation.RuleId).Str("problem", violation.Description).Str("fix", violation.Fix).Send()
	}
}
