package validator

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
)

func ValidateRuleset(ruleset rules.RuleSet, ociTarPath, dockerFilePath string, dockerTarPath string) Violations {
	allowList := getAllowList()
	violations := Violations{}
	for _, rule := range ruleset.Rules {
		if !allowList[rule.Target] {
			continue
		}
		violations.CheckedCount++
		success, fix := rule.Validate(ociTarPath, dockerFilePath, dockerTarPath)
		if success {
			continue
		}
		violations.ViolationCount++
		violation := Violation{
			RuleId: rule.Id,
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

func getAllowList() map[string]bool {
	cfg := config.GetConfig()
	allowList := cfg.TargetList
	if len(allowList) == 0 {
		return map[string]bool{
			"command": true,
			"os":      true,
			"fs":      true,
		}
	}
	log.Debug().Str("allowlist", allowList).Msg("Running only partial targets")
	allowMap := map[string]bool{"command": false,
		"os": false,
		"fs": false,
	}

	allowed := strings.SplitSeq(allowList, ",")
	for target := range allowed {
		if _, ok := allowMap[target]; !ok {
			log.Warn().Str("target", target).Msg("Unknown target in config targetlist")
		}
		allowMap[target] = true
	}

	return allowMap
}
