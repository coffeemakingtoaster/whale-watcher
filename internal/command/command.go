package command

import (
	"fmt"

	"github.com/coffeemakingtoaster/whale-watcher/internal/display"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
	"github.com/rs/zerolog/log"
)

type RunContext struct {
	OCITarballPath    string
	DockerTarballPath string
	DockerFile        string
	RuleSetEntrypoint string
	Instruction       string
	ExportHTML        bool
}

var helpText = `
Whale watcher!
Valid commands:
	- help -> its this one :)
	- validate <ruleset> <dockerfile> <oci image tarball> -> validate the ruleset against the given container artifacts
	- docs <ruleset> -> Serve the ruleset documentation as a website. Pass --export to output and index.html instead
	`

func Run(args []string) int {
	runContext, err := getContext(args)
	if err != nil {
		panic(err)
	}
	if runContext.Instruction == "help" {
		fmt.Println(helpText)
		return 0
	}
	ruleSet, err := rules.LoadRuleset(runContext.RuleSetEntrypoint)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("Loaded %d rules!", len(ruleSet.Rules))
	if runContext.Instruction == "docs" {
		display.ServeRules(ruleSet, runContext.ExportHTML)
		return 0
	}
	// Get ref to prevent directory cleanup
	ref := runner.GetReferencingWorkingDirectoryInstance()
	// Attempt clean exit, force exit if needed
	defer func() {
		if !ref.Free() {
			ref.ForceFree()
		}
	}()

	violations := getViolations(runContext, ruleSet)

	// should a pr be created?
	if config.ShouldInteractWithVSC() {
		err = adapters.CreatePRForFixes(violations, ref.GetAbsolutePath("./Dockerfile"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create PR for changes/fixes")
		}
	} else {
		log.Info().Msg("No git context, no interaction with VSC platform needed")
	}

	if err != nil {
		return 1
	}

	if violations.ViolationCount > 0 {
		return 1
	}
	return 0
}

func getViolations(runContext *RunContext, ruleSet rules.RuleSet) validator.Violations {
	// TODO: These paths are passed down way to far without any validation
	violations := validator.ValidateRuleset(ruleSet, runContext.OCITarballPath, runContext.DockerFile, runContext.DockerTarballPath)
	log.Info().Msgf("Total: %d Violations: %d Fixable: %d", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
	for _, violation := range violations.Violations {
		log.Warn().Str("ruleId", violation.RuleId).Str("problem", violation.Description).Send()
	}
	return violations
}
