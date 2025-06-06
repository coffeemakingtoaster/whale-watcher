package validator

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
}
