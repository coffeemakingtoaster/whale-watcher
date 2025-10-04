package validator

import (
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	violationTypes "github.com/coffeemakingtoaster/whale-watcher/pkg/validator/violations"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func ValidateRuleset(ruleset rules.RuleSet, ociTarPath, dockerFilePath string, dockerTarPath string) violationTypes.Violations {
	violations := violationTypes.Violations{}
	log.Info().Msg(viper.GetString("targetlist"))
	for _, rule := range ruleset.Rules {
		if !config.AllowsTarget(rule.Target) {
			log.Info().Str("id", rule.Id).Msg("Skipped because target is disallowed")
			continue
		}
		violations.CheckedCount++
		success, fix := rule.Validate(ociTarPath, dockerFilePath, dockerTarPath)
		if success {
			continue
		}
		log.Info().Str("id", rule.Id).Msg("Violation detected")
		violations.ViolationCount++
		violation := violationTypes.Violation{
			RuleId:      rule.Id,
			Description: rule.Description,
		}
		if (fix.Fix != "" || rule.FixInstruction != "") && !viper.GetBool("nofix") {
			violations.FixableCount++
			violation.Fix = fix.Fix
			err := rule.PerformFix()
			if err != nil {
				violation.AutoFixed = false
			} else {
				violation.AutoFixed = true
			}
		}
		violations.Violations = append(violations.Violations, violation)
	}
	return violations
}
