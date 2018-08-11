package helm

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	htype "github.com/banzaicloud/banzai-types/components/helm"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/repo"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/pkg/helm"
	yerrors "yunion.io/x/pkg/util/errors"
)

var (
	defaultMan   *HelmManager
	NotInitError error = errors.New("Default helm manager not init")
)

const (
	DefaultTillerImage = "zexi/tiller:v2.9.0"
)

func NewK8sRegionConfig() helm.K8sRegionConfig {
	return helm.NewK8sRegionConfig()
}

func ParseLocalK8sConfig(configPaths []string) (helm.K8sRegionConfig, error) {
	// configPath like: Beijing:kube_config_cluster.yml <REGION_NAME>:xxx.yaml
	if len(configPaths) == 0 {
		return nil, fmt.Errorf("No kubeconfig files provided")
	}
	var confMap helm.K8sRegionConfig
	confMap = make(map[string]string)
	for _, conf := range configPaths {
		baseName := path.Base(conf)
		parts := strings.Split(baseName, "+")
		if len(parts) < 2 {
			log.Errorf("Invalid config path name: %q", conf)
			continue
		}
		regionName := parts[0]
		err := confMap.AddConfig(regionName, conf)
		if err != nil {
			log.Errorf("Add region: %q, config: %q to config map: %v", regionName, conf, err)
			continue
		}
	}
	return confMap, nil
}

type HelmManager struct {
	stateStorePath string
	kubeConfigMap  helm.K8sRegionConfig
}

func NewHelmManager(stateStorePath string, confMap helm.K8sRegionConfig) (*HelmManager, error) {
	err := helm.InitStateStoreDir(stateStorePath)
	if err != nil {
		return nil, err
	}
	m := &HelmManager{
		stateStorePath: stateStorePath,
		kubeConfigMap:  confMap,
	}
	helm.Init(m.kubeConfigMap)
	return m, err
}

func NewHelmManagerByKubeConfs(stateStorePath string, kubePaths []string) (*HelmManager, error) {
	confMap, err := ParseLocalK8sConfig(kubePaths)
	if err != nil {
		return nil, err
	}
	return NewHelmManager(stateStorePath, confMap)
}

func (m *HelmManager) Install(s *mcclient.ClientSession, params jsonutils.JSONObject) error {
	regionName := s.GetRegion()
	return m.InstallByRegion(regionName, params)
}

func (m *HelmManager) InstallToAllRegions() error {
	params := jsonutils.NewDict()
	errs := make([]error, 0)
	for region := range m.kubeConfigMap {
		err := m.InstallByRegion(region, params)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return yerrors.NewAggregate(errs)
}

func (m *HelmManager) KubeConfigBytes(regionName string) ([]byte, error) {
	return m.kubeConfigMap.GetConfigBytes(regionName)
}

func (m *HelmManager) InstallByRegion(regionName string, params jsonutils.JSONObject) error {
	conf, err := m.KubeConfigBytes(regionName)
	if err != nil {
		return err
	}
	if img, _ := params.GetString("tiller_image"); img == "" {
		params.(*jsonutils.JSONDict).Set("tiller_image", jsonutils.NewString(DefaultTillerImage))
	}
	if sa, _ := params.GetString("service_account"); sa == "" {
		params.(*jsonutils.JSONDict).Set("service_account", jsonutils.NewString("tiller"))
	}
	if ns, _ := params.GetString("namespace"); ns == "" {
		params.(*jsonutils.JSONDict).Set("namespace", jsonutils.NewString("kube-system"))
	}
	var opt htype.Install
	err = params.Unmarshal(&opt)
	if err != nil {
		return fmt.Errorf("Unmarshal params %#v error: %v", params, err)
	}
	return helm.Install(&opt, conf, regionName)
}

func (m *HelmManager) TillerStatus(s *mcclient.ClientSession) (string, error) {
	return helm.TillerStatus(s.GetRegion())
}

func (m *HelmManager) ReposList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	repos, err := helm.ReposList(s.GetRegion())
	if err != nil {
		return nil, err
	}
	limit, offset := listQuery(params)
	var res RepoResults = repos
	return NewListResult(res, limit, offset), nil
}

func (m *HelmManager) RepoShow(s *mcclient.ClientSession, name string) (*repo.Entry, error) {
	return helm.RepoShow(name, s.GetRegion())
}

func (m *HelmManager) RepoAdd(s *mcclient.ClientSession, params jsonutils.JSONObject) error {
	var opt repo.Entry
	err := params.Unmarshal(&opt)
	if err != nil {
		return fmt.Errorf("Unmarshal params %#v error: %v", params, err)
	}
	return helm.RepoAdd(s.GetRegion(), &opt)
}

func (m *HelmManager) RepoAddToRegion(regionName string, repos map[string]string) error {
	errs := make([]error, 0)
	for repoName, url := range repos {
		if err := helm.RepoAdd(regionName, &repo.Entry{
			Name: repoName,
			URL:  url,
		}); err != nil {
			errs = append(errs, err)
		}
	}
	return yerrors.NewAggregate(errs)
}

func (m *HelmManager) RepoAddToAllRegions(repos map[string]string) error {
	errs := make([]error, 0)
	for region := range m.kubeConfigMap {
		if err := m.RepoAddToRegion(region, repos); err != nil {
			errs = append(errs, err)
		}
	}
	return yerrors.NewAggregate(errs)
}

func (m *HelmManager) RepoModify(s *mcclient.ClientSession, params jsonutils.JSONObject, repoName string) error {
	var newRepo repo.Entry
	err := params.Unmarshal(&newRepo)
	if err != nil {
		return fmt.Errorf("Unmarshal params %#v error: %v", params, err)
	}
	return helm.RepoModify(s.GetRegion(), repoName, &newRepo)
}

func (m *HelmManager) RepoUpdate(s *mcclient.ClientSession, repos []string) error {
	return helm.ReposUpdate(s.GetRegion(), repos)
}

func (m *HelmManager) RepoDelete(s *mcclient.ClientSession, repoName string) error {
	return helm.RepoDelete(s.GetRegion(), repoName)
}

func (m *HelmManager) ChartsList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	var q helm.ChartQuery
	err := params.Unmarshal(&q)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal params %#v error: %v", params, err)
	}
	charts, err := helm.ChartsList(s.GetRegion(), q)
	if err != nil {
		return nil, err
	}
	var res ChartResults = charts
	sort.Sort(res)
	limit, offset := listQuery(params)
	return NewListResult(res, limit, offset), nil
}

func (m *HelmManager) ChartDetails(s *mcclient.ClientSession, repo, name, version string) (*helm.ChartDetails, error) {
	regionName := s.GetRegion()
	return helm.ChartShowDetails(regionName, repo, name, version)
}

func setBool(params *jsonutils.JSONDict, keys ...string) {
	for _, key := range keys {
		val := jsonutils.JSONFalse
		if jsonutils.QueryBoolean(params, key, false) {
			val = jsonutils.JSONTrue
		}
		params.Set(key, val)
	}
}

func (m *HelmManager) ReleasesList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	regionName := s.GetRegion()
	setBool(params.(*jsonutils.JSONDict), "admin", "all")
	var q helm.ReleaseListQuery
	err := params.Unmarshal(&q)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal params %#v error: %v", params, err)
	}
	if !s.IsSystemAdmin() && q.Admin {
		return nil, fmt.Errorf("%s is not admin", s.GetTenantName())
	}
	if len(q.Namespace) == 0 {
		q.Namespace = s.GetTenantName()
	}
	if q.Admin {
		q.Namespace = ""
	}
	res, err := helm.ReleasesList(q, regionName)
	if err != nil {
		return nil, err
	}
	var releases ReleaseResults = res.GetReleases()
	limit, offset := listQuery(params)
	return NewListResult(releases, limit, offset), nil
}

func IsOwner(s *mcclient.ClientSession, namespace string) bool {
	if s.IsSystemAdmin() {
		return true
	}
	if s.GetTenantName() == namespace {
		return true
	}
	return false
}

func AllowDoReleaseAction(s *mcclient.ClientSession, releaseName string) error {
	regionName := s.GetRegion()
	release, err := helm.ReleaseStatus(releaseName, regionName)
	if err != nil {
		return err
	}
	namespace := release.Namespace
	if !IsOwner(s, namespace) {
		return fmt.Errorf("%q not allowed do action, release: %q, namespace: %q", s.GetTenantName(), releaseName, namespace)
	}
	return nil
}

func (m *HelmManager) ReleaseCreate(s *mcclient.ClientSession, params jsonutils.JSONObject) (*rls.InstallReleaseResponse, error) {
	r, err := helm.NewCreateUpdateReleaseReq(params)
	if err != nil {
		return nil, err
	}
	if r.Namespace == "" {
		r.Namespace = s.GetTenantName()
	}
	if !IsOwner(s, r.Namespace) {
		return nil, fmt.Errorf("%s not allowed create release to namespace: %s", s.GetTenantName(), r.Namespace)
	}
	regionName := s.GetRegion()
	return helm.ReleaseCreate(r, regionName)
}

func (m *HelmManager) ReleaseUpgrade(s *mcclient.ClientSession, params jsonutils.JSONObject) (*rls.UpdateReleaseResponse, error) {
	r, err := helm.NewCreateUpdateReleaseReq(params)
	if err != nil {
		return nil, err
	}
	err = AllowDoReleaseAction(s, r.ReleaseName)
	if err != nil {
		return nil, fmt.Errorf("Upgrade: %v", err)
	}
	return helm.ReleaseUpgrade(r, s.GetRegion())
}

func (m *HelmManager) ReleaseShow(s *mcclient.ClientSession, name string) (*ReleaseDetailsResult, error) {
	regionName := s.GetRegion()
	ret, err := helm.ReleaseShow(name, regionName)
	if err != nil {
		return nil, err
	}
	resp, err := helm.ReleaseStatus(name, regionName)
	if err != nil {
		return nil, err
	}
	rel := ret.GetRelease()
	return NewReleaseDetailsResult(rel, resp.Info, regionName)
}

func (m *HelmManager) ReleaseStatus(s *mcclient.ClientSession, name string) (*ReleaseStatusResult, error) {
	resp, err := helm.ReleaseStatus(name, s.GetRegion())
	if err != nil {
		return nil, err
	}
	resStr := resp.Info.Status.GetResources()
	return &ReleaseStatusResult{
		GetReleaseStatusResponse: resp,
		Resources:                ParseResources(resStr),
	}, nil
}

func (m *HelmManager) ReleaseDelete(s *mcclient.ClientSession, name string) error {
	regionName := s.GetRegion()
	if err := AllowDoReleaseAction(s, name); err != nil {
		return err
	}
	return helm.ReleaseDelete(name, regionName)
}

func Init(stateStorePath string, confMap helm.K8sRegionConfig) error {
	var err error
	defaultMan, err = NewHelmManager(stateStorePath, confMap)
	return err
}

func InitByKubeConfs(stateStorePath string, confs []string) error {
	var err error
	defaultMan, err = NewHelmManagerByKubeConfs(stateStorePath, confs)
	return err
}

func KubeConfigBytes(regionName string) ([]byte, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.KubeConfigBytes(regionName)
}

func Install(s *mcclient.ClientSession, params jsonutils.JSONObject) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.Install(s, params)
}

func InstallByRegion(regionName string, params jsonutils.JSONObject) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.InstallByRegion(regionName, params)
}

func TillerStatus(s *mcclient.ClientSession) (string, error) {
	if defaultMan == nil {
		return "", NotInitError
	}
	return defaultMan.TillerStatus(s)
}

func ReposList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReposList(s, params)
}

func RepoAdd(s *mcclient.ClientSession, params jsonutils.JSONObject) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.RepoAdd(s, params)
}

func RepoShow(s *mcclient.ClientSession, name string) (*repo.Entry, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.RepoShow(s, name)
}

func RepoModify(s *mcclient.ClientSession, params jsonutils.JSONObject, repoName string) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.RepoModify(s, params, repoName)
}

func RepoUpdate(s *mcclient.ClientSession, repos []string) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.RepoUpdate(s, repos)
}

func RepoDelete(s *mcclient.ClientSession, repoName string) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.RepoDelete(s, repoName)
}

func ChartsList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ChartsList(s, params)
}

func ChartDetails(s *mcclient.ClientSession, repo, name, version string) (*helm.ChartDetails, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ChartDetails(s, repo, name, version)
}

func ReleasesList(s *mcclient.ClientSession, params jsonutils.JSONObject) (*ListResult, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReleasesList(s, params)
}

func ReleaseCreate(s *mcclient.ClientSession, params jsonutils.JSONObject) (*rls.InstallReleaseResponse, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReleaseCreate(s, params)
}

func ReleaseUpgrade(s *mcclient.ClientSession, params jsonutils.JSONObject) (*rls.UpdateReleaseResponse, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReleaseUpgrade(s, params)
}

func ReleaseShow(s *mcclient.ClientSession, name string) (*ReleaseDetailsResult, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReleaseShow(s, name)
}

func ReleaseStatus(s *mcclient.ClientSession, name string) (*ReleaseStatusResult, error) {
	if defaultMan == nil {
		return nil, NotInitError
	}
	return defaultMan.ReleaseStatus(s, name)
}

func ReleaseDelete(s *mcclient.ClientSession, name string) error {
	if defaultMan == nil {
		return NotInitError
	}
	return defaultMan.ReleaseDelete(s, name)
}

func InitTillerRepos(repos map[string]string) error {
	if defaultMan == nil {
		return NotInitError
	}
	err := defaultMan.InstallToAllRegions()
	if err != nil && !strings.Contains(err.Error(), "already installed") {
		return err
	}
	return defaultMan.RepoAddToAllRegions(repos)
}
