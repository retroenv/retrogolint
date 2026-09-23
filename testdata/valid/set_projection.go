package valid

import "github.com/retroenv/retrogolib/set"

type orderedValues struct{}

func (orderedValues) ToSlice() []string { return nil }

func projectOrderedValues(values orderedValues, members set.Set[string]) []string {
	_ = set.Sorted(members)
	return values.ToSlice()
}
