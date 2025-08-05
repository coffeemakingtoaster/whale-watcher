package runner_test

import (
	"reflect"
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
)

func TestSliceSearch(t *testing.T) {
	input := []string{"a", "bc", "e", "e", "x", "ee", "e", "e"}
	expected := []int{3, 7}
	actual := []int{}
	search := runner.NewSliceSearch[string]([]string{"e", "e"})
	for i, v := range input {
		if search.Match(v) {
			actual = append(actual, i)
		}
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Index mismatch: Expected %v Got %v", expected, actual)
	}
}
