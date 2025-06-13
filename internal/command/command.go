package command

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/internal/display"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/validator"
)

type RunContext struct {
	OCITarballPath    string
	DockerFile        string
	RuleSetEntrypoint string
	Instruction       string
}

var helpText = `
Whale watcher!
Valid commands:
	- help -> its this one :)
	- validate <ruleset> <dockerfile> <oci image tarball> -> validate the ruleset against the given container artifacts
	- docs <ruleset> -> pretty pring a ruleset
	`

func Run(args []string) {
	runContext, err := getContext(args)
	if err != nil {
		panic(err)
	}
	if runContext.Instruction == "help" {
		fmt.Println(helpText)
		return
	}
	ruleSet, err := rules.LoadRuleset(runContext.RuleSetEntrypoint)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("Loaded %d rules!", len(ruleSet.Rules))
	if runContext.Instruction == "docs" {
		display.PrettyPrintRules(ruleSet, false)
		return
	}
	violations := validator.ValidateRuleset(ruleSet, runContext.OCITarballPath, runContext.DockerFile)
	log.Info().Msgf("Total: %d Violations: %d Fixable: %d", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
	for _, violation := range violations.Violations {
		log.Warn().Str("ruleId", violation.RuleId).Str("problem", violation.Description).Str("fix", violation.Fix).Send()
	}

}
