package validator

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
)

func ValidateRuleset(ruleset rules.RuleSet, imageName, dockerFilePath string) Violations {
	violations := Violations{}
	for _, rule := range ruleset.Rules {
		violations.CheckedCount++
		success, fix := rule.Validate(imageName, dockerFilePath)
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
