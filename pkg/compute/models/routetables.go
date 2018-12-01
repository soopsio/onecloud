package models

import (
	"context"
	"net"
	"reflect"
	"strings"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/gotypes"
	"yunion.io/x/pkg/util/compare"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/validators"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SRoute struct {
	Type        string
	Cidr        string
	NextHopType string
	NextHopId   string
}

func (route *SRoute) Validate(data *jsonutils.JSONDict) error {
	if strings.Index(route.Cidr, "/") > 0 {
		_, ipNet, err := net.ParseCIDR(route.Cidr)
		if err != nil {
			return err
		}
		// normalize from 192.168.1.3/24 to 192.168.1.0/24
		route.Cidr = ipNet.String()
	} else {
		ip := net.ParseIP(route.Cidr).To4()
		if ip == nil {
			return httperrors.NewInputParameterError("invalid addr %s", route.Cidr)
		}
	}
	return nil
}

type SRoutes []*SRoute

func (routes *SRoutes) String() string {
	return jsonutils.Marshal(routes).String()
}
func (routes *SRoutes) IsZero() bool {
	if len([]*SRoute(*routes)) == 0 {
		return true
	}
	return false
}

func (routes *SRoutes) Validate(data *jsonutils.JSONDict) error {
	found := map[string]bool{}
	for _, route := range *routes {
		if err := route.Validate(data); err != nil {
			return err
		}
		if _, ok := found[route.Cidr]; ok {
			// error so that the user has a chance to deal with comments
			return httperrors.NewInputParameterError("duplicate route cidr %s", route.Cidr)
		}
		// TODO aliyun: check overlap with System type route
		found[route.Cidr] = true
	}
	return nil
}

type SRouteTableManager struct {
	db.SVirtualResourceBaseManager
}

var RouteTableManager *SRouteTableManager

func init() {
	gotypes.RegisterSerializable(reflect.TypeOf(&SRoutes{}), func() gotypes.ISerializable {
		return &SRoutes{}
	})
	RouteTableManager = &SRouteTableManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			SRouteTable{},
			"route_tables_tbl",
			"route_table",
			"route_tables",
		),
	}
}

type SRouteTable struct {
	db.SVirtualResourceBase
	SManagedResourceBase

	VpcId         string   `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required"`
	CloudregionId string   `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"`
	Type          string   `width:"16" charset:"ascii" nullable:"false" list:"user"`
	Routes        *SRoutes `list:"user" update:"user" create:"required"`
}

func (man *SRouteTableManager) ListItemFilter(ctx context.Context, q *sqlchemy.SQuery, userCred mcclient.TokenCredential, query jsonutils.JSONObject) (*sqlchemy.SQuery, error) {
	q, err := man.SVirtualResourceBaseManager.ListItemFilter(ctx, q, userCred, query)
	if err != nil {
		return nil, err
	}
	userProjId := userCred.GetProjectId()
	data := query.(*jsonutils.JSONDict)
	for _, key := range []string{"vpc", "cloudregion"} {
		v := validators.NewModelIdOrNameValidator(key, key, userProjId)
		v.Optional(true)
		q, err = v.QueryFilter(q, data)
		if err != nil {
			return nil, err
		}
	}
	return q, nil
}

func (man *SRouteTableManager) validateRoutes(data *jsonutils.JSONDict, update bool) (*jsonutils.JSONDict, error) {
	routes := SRoutes{}
	routesV := validators.NewStructValidator("routes", &routes)
	if update {
		routesV.Optional(true)
	}
	err := routesV.Validate(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (man *SRouteTableManager) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, ownerProjId string, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	data, err := man.validateRoutes(data, false)
	if err != nil {
		return nil, err
	}
	vpcV := validators.NewModelIdOrNameValidator("vpc", "vpc", ownerProjId)
	if err := vpcV.Validate(data); err != nil {
		return nil, err
	}
	vpc := vpcV.Model.(*SVpc)
	cloudregion := vpc.GetRegion()
	if cloudregion == nil {
		return nil, httperrors.NewConflictError("failed fetching cloudregion of vpc %s(%s)", vpc.Name, vpc.Id)
	}
	data.Set("cloudregion_id", jsonutils.NewString(cloudregion.Id))
	return man.SVirtualResourceBaseManager.ValidateCreateData(ctx, userCred, ownerProjId, query, data)
}

func (rt *SRouteTable) ValidateUpdateData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	data, err := RouteTableManager.validateRoutes(data, true)
	if err != nil {
		return nil, err
	}
	return rt.SVirtualResourceBase.ValidateUpdateData(ctx, userCred, query, data)
}

func (rt *SRouteTable) AllowPerformAddRoutes(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) bool {
	return rt.IsOwner(userCred) || userCred.IsSystemAdmin()
}

func (rt *SRouteTable) AllowPerformDelRoutes(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) bool {
	return rt.AllowPerformAddRoutes(ctx, userCred, query, data)
}

// PerformAddRoutes patches acl entries by adding then deleting the specified acls.
// This is intended mainly for command line operations.
func (rt *SRouteTable) PerformAddRoutes(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	routes := gotypes.DeepCopy(rt.Routes).(SRoutes)
	{
		adds := SRoutes{}
		addsV := validators.NewStructValidator("routes", &adds)
		addsV.Optional(true)
		err := addsV.Validate(data)
		if err != nil {
			return nil, err
		}
		for _, add := range adds {
			found := false
			for _, route := range routes {
				if route.Cidr == add.Cidr {
					found = true
					break
				}
			}
			if !found {
				routes = append(routes, add)
			}
		}
	}
	_, err := rt.GetModelManager().TableSpec().Update(rt, func() error {
		rt.Routes = &routes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (rt *SRouteTable) PerformDelRoutes(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	routes := gotypes.DeepCopy(rt.Routes).(SRoutes)
	{
		cidrs := []string{}
		err := data.Unmarshal(&cidrs, "cidrs")
		if err != nil {
			return nil, httperrors.NewInputParameterError("unmarshaling cidrs failed: %s", err)
		}
		for _, cidr := range cidrs {
			for i := len(routes) - 1; i >= 0; i-- {
				route := routes[i]
				if route.Type == "system" {
					continue
				}
				if route.Cidr == cidr {
					routes = append(routes[:i], routes[i+1:]...)
					break
				}
			}
		}
	}
	_, err := rt.GetModelManager().TableSpec().Update(rt, func() error {
		rt.Routes = &routes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (rt *SRouteTable) GetCustomizeColumns(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) *jsonutils.JSONDict {
	extra := rt.SVirtualResourceBase.GetCustomizeColumns(ctx, userCred, query)
	vpcM, err := VpcManager.FetchById(rt.VpcId)
	if err != nil {
		log.Errorf("route table %s(%s): fetch vpc (%s) error: %s",
			rt.Name, rt.Id, rt.VpcId, err)
		return extra
	}
	cloudregionM, err := CloudregionManager.FetchById(rt.CloudregionId)
	if err != nil {
		log.Errorf("route table %s(%s): fetch cloud region (%s) error: %s",
			rt.Name, rt.Id, rt.CloudregionId, err)
		return extra
	}
	extra.Set("vpc", jsonutils.NewString(vpcM.GetName()))
	extra.Set("cloudregion", jsonutils.NewString(cloudregionM.GetName()))
	return extra
}

func (rt *SRouteTable) GetExtraDetails(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) *jsonutils.JSONDict {
	extra := rt.GetCustomizeColumns(ctx, userCred, query)
	extra = rt.SManagedResourceBase.getExtraDetails(ctx, extra)
	return extra
}

func (man *SRouteTableManager) SyncRouteTables(ctx context.Context, userCred mcclient.TokenCredential, vpc *SVpc, cloudRouteTables []cloudprovider.ICloudRouteTable) ([]SRouteTable, []cloudprovider.ICloudRouteTable, compare.SyncResult) {
	localRouteTables := make([]SRouteTable, 0)
	remoteRouteTables := make([]cloudprovider.ICloudRouteTable, 0)
	syncResult := compare.SyncResult{}

	dbRouteTables := []SRouteTable{}
	if err := db.FetchModelObjects(man, man.Query(), &dbRouteTables); err != nil {
		syncResult.Error(err)
		return nil, nil, syncResult
	}
	removed := make([]SRouteTable, 0)
	commondb := make([]SRouteTable, 0)
	commonext := make([]cloudprovider.ICloudRouteTable, 0)
	added := make([]cloudprovider.ICloudRouteTable, 0)
	if err := compare.CompareSets(dbRouteTables, cloudRouteTables, &removed, &commondb, &commonext, &added); err != nil {
		syncResult.Error(err)
		return nil, nil, syncResult
	}

	for i := 0; i < len(commondb); i += 1 {
		err := commondb[i].SyncWithCloudRouteTable(userCred, vpc, commonext[i])
		if err != nil {
			syncResult.UpdateError(err)
			continue
		}
		localRouteTables = append(localRouteTables, commondb[i])
		remoteRouteTables = append(remoteRouteTables, commonext[i])
		syncResult.Update()
	}

	for i := 0; i < len(added); i += 1 {
		routeTableNew, err := man.insertFromCloud(userCred, vpc, added[i])
		if err != nil {
			syncResult.AddError(err)
			continue
		}
		localRouteTables = append(localRouteTables, *routeTableNew)
		remoteRouteTables = append(remoteRouteTables, added[i])
		syncResult.Add()
	}
	return localRouteTables, remoteRouteTables, syncResult
}

func (man *SRouteTableManager) newRouteTableFromCloud(userCred mcclient.TokenCredential, vpc *SVpc, cloudRouteTable cloudprovider.ICloudRouteTable) (*SRouteTable, error) {
	routes := []*SRoute{}
	{
		cloudRoutes, err := cloudRouteTable.GetIRoutes()
		if err != nil {
			return nil, err
		}
		for _, cloudRoute := range cloudRoutes {
			route := &SRoute{
				Type:        cloudRoute.GetType(),
				Cidr:        cloudRoute.GetCidr(),
				NextHopType: cloudRoute.GetNextHopType(),
				NextHopId:   cloudRoute.GetNextHop(),
			}
			routes = append(routes, route)
		}
	}
	routeTable := &SRouteTable{
		CloudregionId: vpc.CloudregionId,
		VpcId:         vpc.Id,
		Type:          cloudRouteTable.GetType(),
		Routes:        (*SRoutes)(&routes),
	}
	routeTable.Name = cloudRouteTable.GetName()
	routeTable.ManagerId = vpc.ManagerId
	routeTable.ExternalId = cloudRouteTable.GetGlobalId()
	routeTable.Description = cloudRouteTable.GetDescription()
	routeTable.ProjectId = userCred.GetProjectId()
	routeTable.SetModelManager(man)
	return routeTable, nil
}

func (man *SRouteTableManager) insertFromCloud(userCred mcclient.TokenCredential, vpc *SVpc, cloudRouteTable cloudprovider.ICloudRouteTable) (*SRouteTable, error) {
	routeTable, err := man.newRouteTableFromCloud(userCred, vpc, cloudRouteTable)
	if err != nil {
		return nil, err
	}
	if err := man.TableSpec().Insert(routeTable); err != nil {
		return nil, err
	}
	return routeTable, nil
}

func (self *SRouteTable) SyncWithCloudRouteTable(userCred mcclient.TokenCredential, vpc *SVpc, cloudRouteTable cloudprovider.ICloudRouteTable) error {
	man := self.GetModelManager().(*SRouteTableManager)
	routeTable, err := man.newRouteTableFromCloud(userCred, vpc, cloudRouteTable)
	if err != nil {
		return err
	}
	_, err = man.TableSpec().Update(self, func() error {
		self.CloudregionId = routeTable.CloudregionId
		self.VpcId = vpc.Id
		self.Type = routeTable.Type
		self.Routes = routeTable.Routes
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}