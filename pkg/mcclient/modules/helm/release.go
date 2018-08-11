package helm

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/gosuri/uitable"
	"github.com/gosuri/uitable/util/strutil"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/timeconv"

	"yunion.io/x/log"
	"yunion.io/x/pkg/helm"
)

var printReleaseTemplate = `REVISION: {{.Release.Version}}
RELEASED: {{.ReleaseDate}}
CHART: {{.Release.Chart.Metadata.Name}}-{{.Release.Chart.Metadata.Version}}
USER-SUPPLIED VALUES:
{{.Release.Config.Raw}}
COMPUTED VALUES:
{{.ComputedValues}}
HOOKS:
{{- range .Release.Hooks }}
---
# {{.Name}}
{{.Manifest}}
{{- end }}
MANIFEST:
{{.Release.Manifest}}
`

func PrintRelease(out io.Writer, rel *release.Release, details bool) error {
	if rel == nil {
		return nil
	}
	fmt.Fprintf(out, "NAME:		%s\n", rel.Name)
	if !details {
		return nil
	}
	cfg, err := chartutil.CoalesceValues(rel.Chart, rel.Config)
	if err != nil {
		return err
	}
	cfgStr, err := cfg.YAML()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Release":        rel,
		"ComputedValues": cfgStr,
		"ReleaseDate":    timeconv.Format(rel.Info.LastDeployed, time.ANSIC),
	}
	return tpl(printReleaseTemplate, data, out)
}

func tpl(t string, vals map[string]interface{}, out io.Writer) error {
	tt, err := template.New("_").Parse(t)
	if err != nil {
		return err
	}
	return tt.Execute(out, vals)
}

func PrintReleaseStatus(out io.Writer, status *ReleaseStatusResult) {
	res := status.GetReleaseStatusResponse
	if res.Info.LastDeployed != nil {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", timeconv.String(res.Info.LastDeployed))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", res.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", res.Info.Status.Code)
	fmt.Fprintf(out, "\n")
	if len(res.Info.Status.Resources) > 0 {
		re := regexp.MustCompile("  +")

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "RESOURCES:\n%s\n", re.ReplaceAllString(res.Info.Status.Resources, "\t"))
		w.Flush()
	}
	if res.Info.Status.LastTestSuiteRun != nil {
		lastRun := res.Info.Status.LastTestSuiteRun
		fmt.Fprintf(out, "TEST SUITE:\n%s\n%s\n\n%s\n",
			fmt.Sprintf("Last Started: %s", timeconv.String(lastRun.StartedAt)),
			fmt.Sprintf("Last Completed: %s", timeconv.String(lastRun.CompletedAt)),
			formatTestResults(lastRun.Results))
	}

	if len(res.Info.Status.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", res.Info.Status.Notes)
	}
}

func formatTestResults(results []*release.TestRun) string {
	tbl := uitable.New()
	tbl.MaxColWidth = 50
	tbl.AddRow("TEST", "STATUS", "INFO", "STARTED", "COMPLETED")
	for i := 0; i < len(results); i++ {
		r := results[i]
		n := r.Name
		s := strutil.PadRight(r.Status.String(), 10, ' ')
		i := r.Info
		ts := timeconv.String(r.StartedAt)
		tc := timeconv.String(r.CompletedAt)
		tbl.AddRow(n, s, i, ts, tc)
	}
	return tbl.String()
}

func MergeValues(values, stringValues []string) ([]byte, error) {
	return helm.MergeValues(values, stringValues)
}

func MergeValuesF(files, values, stringValues []string) ([]byte, error) {
	return helm.MergeValuesF(files, values, stringValues)
}

type ReleaseStatusResult struct {
	*services.GetReleaseStatusResponse
	Resources map[string]Resources `json:"resources"`
}

type ReleaseDetailsResult struct {
	*release.Release
	Resources    map[string]Resources `json:"resources"`
	ConfigValues chartutil.Values     `json:"config_values"`
	RepoURL      string               `json:"repo_url"`
}

func NewReleaseDetailsResult(rel *release.Release, info *release.Info, regionName string) (*ReleaseDetailsResult, error) {
	cfg, err := chartutil.CoalesceValues(rel.Chart, rel.Config)
	if err != nil {
		return nil, fmt.Errorf("CoalesceValues: %v", err)
	}
	r := &ReleaseDetailsResult{
		Release:      rel,
		Resources:    ParseResources(info.Status.GetResources()),
		ConfigValues: cfg,
	}
	url, err := r.GetChartRepoURL(regionName)
	if err != nil {
		log.Errorf("Get chart repo url: %v", err)
	}
	r.RepoURL = url
	return r, nil
}

func (r *ReleaseDetailsResult) GetRelease() *release.Release {
	return r.Release
}

func (r *ReleaseDetailsResult) getChartRepoURL(regionName string) (string, error) {
	name := r.Chart.Metadata.Name
	kws := strings.ToLower(strings.Join(r.Chart.Metadata.Keywords, " "))
	chrts, err := helm.ChartsList(regionName, helm.ChartQuery{
		Name:    name,
		Keyword: kws,
	})
	if err != nil {
		return "", err
	}
	if len(chrts) == 0 {
		err = fmt.Errorf("Not found chart, name: %q", name)
		return "", err
	}
	u, err := url.Parse(chrts[0].Chart.URLs[0])
	if err != nil {
		return "", err
	}
	paths := strings.Split(u.Path, "/")
	u.Path = strings.Join(paths[:len(paths)-1], "/")
	return u.String(), nil
}

func (r *ReleaseDetailsResult) GetChartRepoURL(regionName string) (string, error) {
	ann := r.Chart.Metadata.Annotations
	var url string
	if ann == nil || ann[helm.YunionRepoURLAnnotation] == "" {
		return r.getChartRepoURL(regionName)
	}
	url = ann[helm.YunionRepoURLAnnotation]
	return url, nil
}
