package validator

import (
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/rs/zerolog/log"
)

func ValidateRuleset(ruleset rules.RuleSet, ociTarPath, dockerFilePath string, dockerTarPath string) Violations {
	violations := Violations{}
	cfg := config.GetConfig()
	log.Info().Msg(cfg.TargetList)
	for _, rule := range ruleset.Rules {
		if !cfg.AllowsTarget(rule.Target) {
			log.Info().Str("id", rule.Id).Msg("Skipped because target is disallowed")
			continue
		}
		log.Info().Str("id", rule.Id).Msg("Not skipped")
		violations.CheckedCount++
		success, fix := rule.Validate(ociTarPath, dockerFilePath, dockerTarPath)
		if success {
			continue
		}
		violations.ViolationCount++
		violation := Violation{
			RuleId:      rule.Id,
			Description: rule.Description,
		}
		if fix.Fix != "" || rule.FixInstruction != "" {
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
