package display

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/rules"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// This just assumes color support...if there is none oh well
func PrettyPrintRules(ruleSet rules.RuleSet, colored bool) {
	printfln("Ruleset: %s (Size: %d)", ruleSet.Name, len(ruleSet.Rules))
	for _, rule := range ruleSet.Rules {
		printfln("\t* %s%s: %s%s%s (%s,%s,%s)%s", Blue, rule.Id, Green, rule.Description, Magenta, rule.Category, rule.Scope, rule.Target, Reset)
	}
}
