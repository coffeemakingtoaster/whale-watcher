package display

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
)

func PrettyPrintRules(ruleSet rules.RuleSet, colored bool) {
	printfln("Ruleset: %s (Size: %d)", ruleSet.Name, len(ruleSet.Rules))
	for _, rule := range ruleSet.Rules {
		printfln("* %s: (%s,%s,%s)", rule.Id, rule.Category, rule.Scope, rule.Target)
		printfln("\t%s", rule.Description)
		printfln("\t[%s]", rule.Instruction)
	}
}
