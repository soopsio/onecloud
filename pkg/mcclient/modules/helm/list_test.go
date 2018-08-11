package helm

import (
	"reflect"
	"testing"
)

type IntList []int

func (l IntList) Total() int64 {
	return int64(len(l))
}

func (l IntList) Offset(offset int64) Lister {
	var res IntList = []int{}
	if l.Total() > offset {
		res = l[offset:]
	}
	return res
}

func (l IntList) Range(begin, end int64) Lister {
	return l[begin:end]
}

func (l IntList) Index(i int64) interface{} {
	return l[i]
}

func (l IntList) Columns() []interface{} {
	return []interface{}{"Index", "Number"}
}

func (l IntList) RowKeys(obj interface{}) []interface{} {
	i := obj.(int)
	return []interface{}{"idx", i}
}

func TestListPart(t *testing.T) {
	type args struct {
		list   Lister
		limit  int64
		offset int64
	}
	tests := []struct {
		name string
		args args
		want Lister
	}{
		{
			name: "Limit work",
			args: args{
				list:   IntList{1, 2, 3},
				limit:  2,
				offset: 0,
			},
			want: IntList{1, 2},
		},
		{
			name: "Offset work",
			args: args{
				list:   IntList{1, 2, 3},
				limit:  3,
				offset: 1,
			},
			want: IntList{2, 3},
		},
		{
			name: "Offset > Total",
			args: args{
				list:   IntList{1, 2, 3},
				limit:  3,
				offset: 3,
			},
			want: IntList{},
		},
		{
			name: "Empty lister",
			args: args{
				list:   IntList{},
				limit:  3,
				offset: 3,
			},
			want: IntList{},
		},
		{
			name: "Offset > Limit",
			args: args{
				list:   IntList{1, 2, 3, 4, 5, 6},
				limit:  2,
				offset: 3,
			},
			want: IntList{4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListPart(tt.args.list, tt.args.limit, tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListerTable(t *testing.T) {
	t.Logf("%s", ListerTable(&ListResult{IntList{}, 0, 20, 2}))
	t.Logf("%s", ListerTable(&ListResult{IntList{1, 2, 3}, 3, 20, 0}))
	t.Logf("%s", ListerTable(&ListResult{IntList{1, 2, 3, 4}, 4, 20, 2}))
	t.Logf("%s", ListerTable(&ListResult{IntList{1, 2, 3, 4, 5}, 5, 1, 2}))
}
