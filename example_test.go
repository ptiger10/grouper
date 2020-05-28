package grouper_test

import (
	"fmt"

	"github.com/ptiger10/grouper"
)

func Example_groupReduce() {
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
		func(slice interface{}) interface{} {
			var sum int
			arr := slice.([]record)
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
