package validator

import (
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	violationTypes "github.com/coffeemakingtoaster/whale-watcher/pkg/validator/violations"
	"github.com/spf13/viper"
)

func ValidateRuleset(ruleset rules.RuleSet, ociTarPath, dockerFilePath string, dockerTarPath string) violationTypes.Violations {
	violations := violationTypes.Violations{}

	runner := runner.NewPythonRunner()

	res, err := runner.Run(ruleset, ociTarPath, dockerFilePath, dockerTarPath)
	if err != nil {
		panic(err)
	}

	for _, rule := range ruleset.Rules {
		violations.CheckedCount++

		if res[rule.Id].Success {
			continue
		}

		violation := violationTypes.Violation{
			RuleId:      rule.Id,
			Description: rule.Description,
		}

		if viper.GetBool("no_fix") && len(rule.FixInstruction) == 0 {
			violations.FixableCount++
			violation.AutoFixed = res[rule.Id].Autofix
		}
		violations.Violations = append(violations.Violations, violation)
	}

	return violations
}
