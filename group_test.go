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
			[]string{"bar", "foo"},
			[][]int{
				{1, 2},
				{0, 3},
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
			[]string{"bar", "foo"},
			[][]int{
				{1, 2},
				{0, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Grouper{
				sliceOfStructs: tt.fields.sliceOfStructs,
			}
			gotGroups, gotIndices := g.GroupBy(tt.args.lambda)
			if !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("Grouper.GroupBy() gotGroups = %v, want %v", gotGroups, tt.wantGroups)
			}
			if !reflect.DeepEqual(gotIndices, tt.wantIndices) {
				t.Errorf("Grouper.GroupBy() gotIndices = %v, want %v", gotIndices, tt.wantIndices)
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
		lambda  func(sliceOfStructs interface{}) interface{}
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
				func(sliceOfStructs interface{}) interface{} {
					var sum int
					arr := sliceOfStructs.([]testStructPrivate)
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
			if got := g.Reduce(tt.args.groups, tt.args.indices, tt.args.lambda); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Grouper.Reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGrouper(t *testing.T) {
	foo := "foo"
	type args struct {
		sliceOfStructs interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    Grouper
		wantErr bool
	}{
		{"pass",
			args{
				[]testStructPrivate{
					{"foo", 1},
				},
			},
			Grouper{
				typ: reflect.TypeOf(testStructPrivate{}),
				sliceOfStructs: []testStructPrivate{
					{"foo", 1},
				},
			},
			false,
		},
		{"fail - not slice",
			args{"foo"},
			Grouper{},
			true,
		},
		{"fail - not slice of struct",
			args{[]string{"foo"}},
			Grouper{},
			true,
		},
		{"fail - not slice of *struct",
			args{[]*string{&foo}},
			Grouper{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGrouper(tt.args.sliceOfStructs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGrouper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGrouper() = %v, want %v", got, tt.want)
			}
		})
	}
}
