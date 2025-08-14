package rules_test

import (
	"fmt"
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
)

func TestSwallowNoOverlap(t *testing.T) {
	strongSet := rules.RuleSet{
		Name: "strong",
		Rules: []*rules.Rule{
			{
				Id: "unique_1",
			},
			{
				Id: "unique_2",
			},
		},
	}
	weakSet := rules.RuleSet{
		Name: "weak",
		Rules: []*rules.Rule{
			{
				Id: "unique_3",
			},
			{
				Id: "unique_4",
			},
		},
	}
	strongSet.Swallow(weakSet)

	if len(strongSet.Rules) != 4 {
		t.Errorf("rule count mismatch: Expected 4 Got %d", len(strongSet.Rules))
	}
	for i := range 4 {
		if strongSet.Rules[i].Id != fmt.Sprintf("unique_%d", i+1) {
			t.Errorf("rule mismatch: Expected %s Got %s", fmt.Sprintf("unique_%d", i+1), strongSet.Rules[i].Id)
		}
	}
}

func TestSwallowPartialOverlap(t *testing.T) {
	strongSet := rules.RuleSet{
		Name: "strong",
		Rules: []*rules.Rule{
			{
				Id: "unique_1",
			},
			{
				Id: "unique_2",
			},
			{
				Id:          "unique_3",
				Description: "strong",
			},
		},
	}
	weakSet := rules.RuleSet{
		Name: "weak",
		Rules: []*rules.Rule{
			{
				Id:          "unique_3",
				Description: "weak",
			},
			{
				Id: "unique_4",
			},
		},
	}
	strongSet.Swallow(weakSet)

	if len(strongSet.Rules) != 4 {
		t.Errorf("rule count mismatch: Expected 4 Got %d", len(strongSet.Rules))
	}
	for i := range 4 {
		if strongSet.Rules[i].Id != fmt.Sprintf("unique_%d", i+1) {
			t.Errorf("rule mismatch: Expected %s Got %s", fmt.Sprintf("unique_%d", i+1), strongSet.Rules[i].Id)
		}
	}
	if strongSet.Rules[2].Description != "strong" {
		t.Errorf("Post swallow description mismatch: Expected strong Got %s", strongSet.Rules[2].Description)
	}
}

func TestSwallowFullOverlap(t *testing.T) {
	strongSet := rules.RuleSet{
		Name: "strong",
		Rules: []*rules.Rule{
			{
				Id:          "unique_1",
				Description: "strong",
			},
			{
				Id:          "unique_2",
				Description: "strong",
			},
		},
	}
	weakSet := rules.RuleSet{
		Name: "weak",
		Rules: []*rules.Rule{
			{
				Id:          "unique_1",
				Description: "weak",
			},
			{
				Id:          "unique_2",
				Description: "weak",
			},
		},
	}
	strongSet.Swallow(weakSet)

	if len(strongSet.Rules) != 2 {
		t.Errorf("rule count mismatch: Expected 2 Got %d", len(strongSet.Rules))
	}
	for i := range 2 {
		if strongSet.Rules[i].Id != fmt.Sprintf("unique_%d", i+1) {
			t.Errorf("rule mismatch: Expected %s Got %s", fmt.Sprintf("unique_%d", i+1), strongSet.Rules[i].Id)
		}
		if strongSet.Rules[i].Description != "strong" {
			t.Errorf("Post swallow description mismatch: Expected strong Got %s", strongSet.Rules[i].Description)
		}
	}
}
