// Package grouper provides a data structure (Grouper) and methods
// for splitting a slice of structs into "groups" (smaller slices of structs)
// and reducing the values in each group.
package grouper

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/pkg/errors"
)

// A Grouper can
type Grouper struct {
	sliceOfStructs interface{}
	typ            reflect.Type
}

// New constructs a new Grouper from a slice of structs (or pointers to structs).
func New(sliceOfStructs interface{}) (Grouper, error) {
	err := fmt.Sprintf("unsupported input type (%v), must be []*struct or []struct",
		reflect.TypeOf(sliceOfStructs))
	if reflect.TypeOf(sliceOfStructs).Kind() != reflect.Slice {
		return Grouper{}, errors.New(err)
	}
	typ := reflect.TypeOf(sliceOfStructs).Elem()
	switch typ.Kind() {
	case reflect.Ptr:
		if typ.Elem().Kind() != reflect.Struct {
			return Grouper{}, errors.New(err)
		}

	case reflect.Struct:
	default:
		return Grouper{}, errors.New(err)
	}
	return Grouper{
		sliceOfStructs: sliceOfStructs,
		typ:            typ,
	}, nil
}

// GroupBy assigns each struct to a group based on the grouper function.
func (g Grouper) GroupBy(grouper func(strct interface{}) string) (groupNames []string, indices [][]int) {
	v := reflect.ValueOf(g.sliceOfStructs)
	m := make(map[string][]int, 7)
	for i := 0; i < v.Len(); i++ {
		group := grouper(v.Index(i).Interface())
		if _, ok := m[group]; !ok {
			m[group] = []int{i}
		} else {
			m[group] = append(m[group], i)
		}
	}
	groupNames = make([]string, 0, len(m))
	for k := range m {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)

	indices = make([][]int, len(m))
	for i := range groupNames {
		indices[i] = m[groupNames[i]]
	}
	return groupNames, indices
}

// Reduce reduces one or more groups (slices of structs) to one value per group.
// The result is returned in the form: map{groupName: value}
func (g Grouper) Reduce(
	groupNames []string,
	indices [][]int,
	reducer func(sliceOfStructs interface{}) interface{},
) map[string]interface{} {
	ret := make(map[string]interface{}, len(groupNames))
	for i, group := range groupNames {
		subset := reflect.MakeSlice(reflect.SliceOf(g.typ), len(indices[i]), len(indices[i]))
		for j, index := range indices[i] {
			dst := subset.Index(j)
			src := reflect.ValueOf(g.sliceOfStructs).Index(index)
			dst.Set(src)
		}
		ret[group] = reducer(subset.Interface())
	}
	return ret
}

// GroupReduce calls .Group() followed by .Reduce() on a slice of structs.
func (g Grouper) GroupReduce(
	grouper func(strct interface{}) string,
	reducer func(sliceOfStructs interface{}) interface{},
) map[string]interface{} {
	groups, indices := g.GroupBy(grouper)
	return g.Reduce(groups, indices, reducer)
}
