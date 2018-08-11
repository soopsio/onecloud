package helm

import (
	sysjson "encoding/json"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/gosuri/uitable"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/timeconv"

	json "yunion.io/x/jsonutils"
	"yunion.io/x/pkg/helm"
)

var (
	DefaultLimit int64 = 100
)

func listQuery(params json.JSONObject) (limit, offset int64) {
	limit, _ = params.(*json.JSONDict).Int("limit")
	if limit <= 0 {
		limit = DefaultLimit
	}
	offset, _ = params.(*json.JSONDict).Int("offset")
	if offset < 0 {
		offset = 0
	}
	return
}

type Lister interface {
	Total() int64
	Offset(int64) Lister
	Range(begin, end int64) Lister
	Index(int64) interface{}
}

type ListResult struct {
	Data   Lister
	Total  int64
	Limit  int64
	Offset int64
}

func (r *ListResult) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{}
	alias["data"] = r.Data
	alias["total"] = r.Total
	alias["limit"] = r.Limit
	alias["offset"] = r.Offset
	return sysjson.Marshal(alias)
}

func ListPart(list Lister, limit, offset int64) Lister {
	rest := list.Offset(offset)
	if rest.Total() > limit {
		return rest.Range(0, limit)
	}
	return rest
}

func NewListResult(list Lister, limit, offset int64) *ListResult {
	return &ListResult{
		Data:   ListPart(list, limit, offset),
		Total:  list.Total(),
		Limit:  limit,
		Offset: offset,
	}
}

type PrintLister interface {
	Lister
	Columns() []interface{}
	RowKeys(obj interface{}) []interface{}
}

func ListerTable(res *ListResult) *uitable.Table {
	l := res.Data.(PrintLister)
	table := uitable.New()
	table.MaxColWidth = 80
	//table.Wrap = true
	table.AddRow(l.Columns()...)
	var idx int64
	for ; idx < l.Total(); idx++ {
		table.AddRow(l.RowKeys(l.Index(idx))...)
	}

	return table
}

func PrintListResult(res *ListResult) {
	fmt.Println(ListerTable(res))

	table := uitable.New()
	total := res.Total
	offset := res.Offset
	limit := res.Limit
	page := (offset / limit) + 1
	pages := total / limit
	if pages*limit < total {
		pages += 1
	}
	table.AddRow("")
	table.AddRow("Total", "Pages", "Limit", "Offset", "Page")
	table.AddRow(total, pages, limit, offset, page)
	fmt.Println(table)
}

type RepoResults []*repo.Entry

func (c RepoResults) Total() int64 {
	return int64(len(c))
}

func (c RepoResults) Index(i int64) interface{} {
	return c[i]
}

func (c RepoResults) Offset(offset int64) Lister {
	var res RepoResults
	res = []*repo.Entry{}
	if c.Total() > offset {
		res = c[offset:]
	}
	return res
}

func (c RepoResults) Range(b, e int64) Lister {
	return c[b:e]
}

func (c RepoResults) Columns() []interface{} {
	return []interface{}{"NAME", "URL", "CACHE"}
}

func (c RepoResults) RowKeys(obj interface{}) []interface{} {
	r := obj.(*repo.Entry)
	return []interface{}{r.Name, r.URL, r.Cache}
}

type ReleaseResults []*release.Release

func (c ReleaseResults) Total() int64 {
	return int64(len(c))
}

func (c ReleaseResults) Index(i int64) interface{} {
	return c[i]
}

func (c ReleaseResults) Offset(offset int64) Lister {
	var res ReleaseResults
	res = []*release.Release{}
	if c.Total() > offset {
		res = c[offset:]
	}
	return res
}

func (c ReleaseResults) Range(b, e int64) Lister {
	return c[b:e]
}

func (c ReleaseResults) Columns() []interface{} {
	return []interface{}{"NAME", "REVISION", "UPDATED", "STATUS", "CHART", "NAMESPACE"}
}

func (rc ReleaseResults) RowKeys(obj interface{}) []interface{} {
	r := obj.(*release.Release)
	md := r.GetChart().GetMetadata()
	c := fmt.Sprintf("%s-%s", md.GetName(), md.GetVersion())
	t := "-"
	if tspb := r.GetInfo().GetLastDeployed(); tspb != nil {
		t = timeconv.String(tspb)
	}
	s := r.GetInfo().GetStatus().GetCode().String()
	v := r.GetVersion()
	n := r.GetNamespace()
	return []interface{}{r.GetName(), v, t, s, c, n}
}

type ChartResults []*helm.ChartResult

func (c ChartResults) Total() int64 {
	return int64(len(c))
}

func (c ChartResults) Index(i int64) interface{} {
	return c[i]
}

func (c ChartResults) Offset(offset int64) Lister {
	var res ChartResults
	res = []*helm.ChartResult{}
	if c.Total() > offset {
		res = c[offset:]
	}
	return res
}

func (c ChartResults) Range(begin, end int64) Lister {
	return c[begin:end]
}

func (c ChartResults) Columns() []interface{} {
	return []interface{}{"REPO/NAME", "VERSION", "DESCRIPTION"}
}

func (c ChartResults) RowKeys(obj interface{}) []interface{} {
	ch := obj.(*helm.ChartResult)
	return []interface{}{fmt.Sprintf("%s/%s", ch.Repo, ch.Chart.Name), ch.Chart.Version, ch.Chart.Description}
}

func (c ChartResults) Len() int { return int(c.Total()) }

func (c ChartResults) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c ChartResults) Less(a, b int) bool {
	first := c[a]
	second := c[b]

	name1 := fmt.Sprintf("%s/%s", first.Repo, first.Chart.Name)
	name2 := fmt.Sprintf("%s/%s", second.Repo, second.Chart.Name)

	if name1 == name2 {
		v1, err := semver.NewVersion(first.Chart.Version)
		if err != nil {
			return true
		}
		v2, err := semver.NewVersion(second.Chart.Version)
		if err != nil {
			return true
		}
		// Sort so that the newest chart is higher then the oldest chart.
		return v1.GreaterThan(v2)
	}

	return name1 < name2
}
