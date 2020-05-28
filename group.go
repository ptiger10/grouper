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
// Group names can be accessed afterwards by calling g.Group().
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

// Reduce reduces one or more groups (i.e., slice(s) of structs derived from a larger slice of structs) to one value per group.
// The result is returned in the form: map{groupName: value}
func (g *Grouper) Reduce(
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
func (g *Grouper) GroupReduce(
	grouper func(strct interface{}) string,
	reducer func(sliceOfStructs interface{}) interface{},
) map[string]interface{} {
	indices := g.GroupBy(grouper)
	return g.Reduce(g.Groups(), indices, reducer)
}
