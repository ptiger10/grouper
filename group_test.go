package grouper

import (
	"reflect"
	"testing"
)

type testStructPrivate struct {
	name string
	age  int
}

func TestGrouper_GroupBy(t *testing.T) {
	type fields struct {
		sliceOfStructs interface{}
	}
	type args struct {
		lambda func(scalarStruct interface{}) string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantGroups  []string
		wantIndices [][]int
	}{
		{"pass",
			fields{
				[]testStructPrivate{
					{"foo", 1},
					{"bar", 3},
					{"bar", 5},
					{"foo", 2},
				},
			},
			args{
				func(v interface{}) string {
					return v.(testStructPrivate).name
				},
			},
			[]string{"foo", "bar"},
			[][]int{
				{0, 3},
				{1, 2},
			},
		},
		{"pass - *[]",
			fields{
				[]*testStructPrivate{
					{"foo", 1},
					{"bar", 3},
					{"bar", 5},
					{"foo", 2},
				},
			},
			args{
				func(v interface{}) string {
					return v.(*testStructPrivate).name
				},
			},
			[]string{"foo", "bar"},
			[][]int{
				{0, 3},
				{1, 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Grouper{
				sliceOfStructs: tt.fields.sliceOfStructs,
			}
			gotIndices := g.GroupBy(tt.args.lambda)
			if !reflect.DeepEqual(gotIndices, tt.wantIndices) {
				t.Errorf("Grouper.GroupBy() gotIndices = %v, want %v", gotIndices, tt.wantIndices)
			}
			if !reflect.DeepEqual(g.Groups(), tt.wantGroups) {
				t.Errorf("Grouper.GroupBy() gotGroups = %v, want %v", g.Groups(), tt.wantGroups)
			}
		})
	}
}

func TestGrouper_Reduce(t *testing.T) {
	type fields struct {
		sliceOfStructs interface{}
		typ            reflect.Type
	}
	type args struct {
		groups  []string
		indices [][]int
		reducer func(sliceOfStructs interface{}, name string) interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{"pass",
			fields{
				[]testStructPrivate{
					{"foo", 1},
					{"bar", 3},
					{"bar", 5},
					{"foo", 2},
				},
				reflect.TypeOf(testStructPrivate{}),
			},
			args{
				[]string{"bar", "foo"},
				[][]int{
					{1, 2},
					{0, 3},
				},
				func(slice interface{}, _ string) interface{} {
					var sum int
					arr := slice.([]testStructPrivate)
					for i := range arr {
						sum += arr[i].age
					}
					return sum
				},
			},
			map[string]interface{}{"bar": 8, "foo": 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Grouper{
				sliceOfStructs: tt.fields.sliceOfStructs,
				typ:            tt.fields.typ,
			}
			if got := g.Reduce(tt.args.groups, tt.args.indices, tt.args.reducer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Grouper.Reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	foo := "foo"
	type args struct {
		sliceOfStructs interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *Grouper
		wantErr bool
	}{
		{"pass",
			args{
				[]testStructPrivate{
					{"foo", 1},
				},
			},
			&Grouper{
				typ: reflect.TypeOf(testStructPrivate{}),
				sliceOfStructs: []testStructPrivate{
					{"foo", 1},
				},
			},
			false,
		},
		{"fail - not slice",
			args{"foo"},
			nil,
			true,
		},
		{"fail - not slice of struct",
			args{[]string{"foo"}},
			nil,
			true,
		},
		{"fail - not slice of *struct",
			args{[]*string{&foo}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.sliceOfStructs)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
