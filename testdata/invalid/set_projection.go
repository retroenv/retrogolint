package invalid

import "github.com/retroenv/retrogolib/set"

type setHolder struct {
	members set.Set[string]
}

func newMembers() set.Set[string] {
	return set.New[string]()
}

func projectUnorderedValues(members set.Set[string]) []string {
	return members.ToSlice()
}

func projectStoredValues(holder setHolder) []string {
	return holder.members.ToSlice()
}

func projectReturnedValues() []string {
	return newMembers().ToSlice()
}
