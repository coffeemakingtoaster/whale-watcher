package validator

import (
	"fmt"
	"strings"
)

type Violations struct {
	CheckedCount   int
	ViolationCount int
	FixableCount   int
	Violations     []Violation
}

type Violation struct {
	RuleId      string
	Description string
	Fix         string
	AutoFixed   bool
}

func (v *Violations) BuildDescriptionMarkdown() string {
	var fixed []string
	var detected []string
	var sb strings.Builder

	for _, violation := range v.Violations {
		markdown := buildViolationMarkdown(violation)
		if violation.AutoFixed {
			fixed = append(fixed, markdown)
		} else {
			detected = append(detected, markdown)
		}
	}

	if len(fixed) > 0 {
		sb.WriteString("## ✅ Automatically Fixed Issues\n\n")
		addList(&sb, fixed)
		sb.WriteString("\n")
	}

	if len(detected) > 0 {
		sb.WriteString("## ❌ Detected but Not Automatically Fixable\n\n")
		addList(&sb, detected)
		sb.WriteString("\n")
	}

	return sb.String()
}

func buildViolationMarkdown(v Violation) string {
	return fmt.Sprintf("`%s` - %s", v.RuleId, v.Description)
}

func addList(sb *strings.Builder, elements []string) {
	for _, element := range elements {
		sb.WriteString(fmt.Sprintf("- %s\n", element))
	}
}
