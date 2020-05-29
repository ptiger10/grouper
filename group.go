// Package grouper provides a data structure (Grouper) and methods
// for splitting a slice of structs into "groups" (smaller slices of structs)
// and reducing the values in each group.
package grouper

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// A Grouper holds a slice of structs.
type Grouper struct {
	sliceOfStructs interface{}
	typ            reflect.Type
	groups         []string
}

// Groups returns the group names saved during the most recent GroupBy call,
// in the order in which each group was observed.
func (g *Grouper) Groups() []string {
	return g.groups
}

// New constructs a new Grouper from a slice of structs (or a slice of pointers to structs).
func New(sliceOfStructs interface{}) (*Grouper, error) {
	err := fmt.Sprintf("unsupported input type (%v), must be []*struct or []struct",
		reflect.TypeOf(sliceOfStructs))
	if reflect.TypeOf(sliceOfStructs).Kind() != reflect.Slice {
		return nil, errors.New(err)
	}
	typ := reflect.TypeOf(sliceOfStructs).Elem()
	switch typ.Kind() {
	case reflect.Ptr:
		if typ.Elem().Kind() != reflect.Struct {
			return nil, errors.New(err)
		}

	case reflect.Struct:
	default:
		return nil, errors.New(err)
	}
	return &Grouper{
		sliceOfStructs: sliceOfStructs,
		typ:            typ,
	}, nil
}

// GroupBy assigns each struct to a group based on the grouper function. Returns the index positions of each group.
// Group names can be accessed afterwards by calling .Group().
func (g *Grouper) GroupBy(grouper func(strct interface{}) string) [][]int {
	v := reflect.ValueOf(g.sliceOfStructs)
	m := make(map[string][]int, 7)
	groupNames := make([]string, 0, 7)
	for i := 0; i < v.Len(); i++ {
		group := grouper(v.Index(i).Interface())
		if _, ok := m[group]; !ok {
			m[group] = []int{i}
			groupNames = append(groupNames, group)
		} else {
			m[group] = append(m[group], i)
		}
	}
	g.groups = groupNames

	indices := make([][]int, len(m))
	for i := range groupNames {
		indices[i] = m[groupNames[i]]
	}
	return indices
}

// Reduce uses a reducer function to reduce one or more groups
// (i.e., slice(s) of structs derived from a larger slice of structs) to one interface value per group.
// Returns a map in the form {groupName: reducedValue}.
func (g *Grouper) Reduce(
	indices [][]int,
	reducer func(groupSlice interface{}) interface{},
) map[string]interface{} {
	ret := make(map[string]interface{}, len(g.groups))
	for i, group := range g.groups {
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

// ReduceWithName uses a reducer function to reduce one or more groups
// (i.e., slice(s) of structs derived from a larger slice of structs) to one value per group.
// The reducer function has access to the name of each group.
// The caller is expected to handle the results outside of the reducer function, and so this is a void method.
func (g *Grouper) ReduceWithName(
	indices [][]int,
	reducer func(groupSlice interface{}, name string),
) {
	for i, group := range g.groups {
		subset := reflect.MakeSlice(reflect.SliceOf(g.typ), len(indices[i]), len(indices[i]))
		for j, index := range indices[i] {
			dst := subset.Index(j)
			src := reflect.ValueOf(g.sliceOfStructs).Index(index)
			dst.Set(src)
		}
		reducer(subset.Interface(), group)
	}
	return
}

// GroupReduce calls .Group() followed by .Reduce() on a slice of structs.
func (g *Grouper) GroupReduce(
	grouper func(strct interface{}) string,
	reducer func(groupSlice interface{}) interface{},
) map[string]interface{} {
	indices := g.GroupBy(grouper)
	return g.Reduce(indices, reducer)
}

// GroupReduceWithName calls .Group() followed by .ReduceWithName() on a slice of structs.
func (g *Grouper) GroupReduceWithName(
	grouper func(strct interface{}) string,
	reducer func(groupSlice interface{}, name string),
) {
	indices := g.GroupBy(grouper)
	g.ReduceWithName(indices, reducer)
	return
}
