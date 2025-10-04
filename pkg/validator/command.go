package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	violationTypes "github.com/coffeemakingtoaster/whale-watcher/pkg/validator/violations"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type ValidateContext struct {
	OCITarballPath    string
	DockerTarballPath string
	DockerFilePath    string
	RuleSetEntrypoint string
}

func buildContext(input []string) *ValidateContext {
	if len(input) < 4 {
		input = append(input, make([]string, 4-len(input))...)
	}

	return &ValidateContext{
		RuleSetEntrypoint: input[0],
		DockerFilePath:    input[1],
		OCITarballPath:    input[2],
		DockerTarballPath: input[3],
	}
}

func NewCommand() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate the given inputs based on the policy set",
		Long: `Given a policy sets and input files, validate each policy. 

Expected arguments:  <policy set location> <Dockerfile location> [<oci tar location>] [<docker tar location>]
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Validate needs at least a set policy set and Dockerfile (Got: '%s')", strings.Join(args, " "))
			}
			if len(args) > 4 {
				return fmt.Errorf("Validate only accepts 4 arguments (policy set, Dockerfile, oci tar, docker tar) (Got: '%s')", strings.Join(args, " "))
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := buildContext(args)
			ruleSet, err := rules.LoadRuleset(ctx.RuleSetEntrypoint)
			if err != nil {
				panic(err)
			}
			// Fail code if violations were detected
			if validate(ctx, ruleSet) {
				os.Exit(0)
			}
			os.Exit(1)
		},
	}
	return cmd
}

func validate(ctx *ValidateContext, ruleSet rules.RuleSet) bool {
	var err error
	// Get ref to prevent directory cleanup
	ref := runner.GetReferencingWorkingDirectoryInstance()
	// Attempt clean exit, force exit if needed
	defer func() {
		if !ref.Free() {
			ref.ForceFree()
		}
	}()

	violations := getViolations(ctx, ruleSet)

	if config.ShouldInteractWithVSC() {
		err = adapters.CreatePRForFixes(violations, ref.GetAbsolutePath("./Dockerfile"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create PR for changes/fixes")
		}
	} else {
		log.Info().Msg("No git context, no interaction with VSC platform needed")
	}

	if err != nil {
		return false
	}

	if violations.ViolationCount > 0 {
		return false
	}
	return true
}

func getViolations(runContext *ValidateContext, ruleSet rules.RuleSet) violationTypes.Violations {
	// TODO: These paths are passed down way to far without any validation
	violations := ValidateRuleset(ruleSet, runContext.OCITarballPath, runContext.DockerFilePath, runContext.DockerTarballPath)
	log.Info().Msgf("Total: %d Violations: %d Fixable: %d", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
	for _, violation := range violations.Violations {
		log.Warn().Str("ruleId", violation.RuleId).Str("problem", violation.Description).Send()
	}
	return violations
}
