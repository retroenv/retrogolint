package invalid

var counts = map[string]int{"first": 1, "second": 2}

var countGroups = []map[string]int{
	{"first": 1, "second": 2},
}

var nestedCounts = map[string]map[string]int{
	"group": {"first": 1, "second": 2},
}
