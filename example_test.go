package grouper_test

import (
	"fmt"

	"github.com/ptiger10/grouper"
)

func ExampleGrouper_GroupReduce() {
	type record struct {
		name string
		age  int
	}
	records := []record{
		{"foo", 1},
		{"bar", 3},
		{"bar", 5},
		{"foo", 2},
	}
	g, _ := grouper.New(records)
	results := g.GroupReduce(
		func(v interface{}) string { return v.(record).name },
		func(groupSlice interface{}) interface{} {
			var sum int
			arr := groupSlice.([]record)
			for i := range arr {
				sum += arr[i].age
			}
			return sum
		},
	)
	fmt.Println(results)
	// Output:
	// map[bar:8 foo:3]
}

func ExampleGrouper_GroupReduceWithName() {
	type record struct {
		name string
		age  int
	}
	records := []record{
		{"foo", 1},
		{"bar", 3},
		{"bar", 5},
		{"foo", 2},
	}
	g, _ := grouper.New(records)
	m := make(map[string]int)
	g.GroupReduceWithName(
		func(v interface{}) string { return v.(record).name },
		func(groupSlice interface{}, name string) {
			var sum int
			arr := groupSlice.([]record)
			for i := range arr {
				sum += arr[i].age
			}
			m[name] = sum
		},
	)
	fmt.Println(m)
	// Output:
	// map[bar:8 foo:3]
}
