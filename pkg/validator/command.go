package validator

import (
	"fmt"
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

	// validateCmd represents the validate command
	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
			ruleSet, err := rules.LoadRuleset(ctx.DockerFilePath)
			if err != nil {
				panic(err)
			}
			validate(ctx, ruleSet)
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
