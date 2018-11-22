package sshkeys

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/appctx"
	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
)

func AddSshKeysHandler(prefix string, app *appsrv.Application) {
	app.AddHandler2("GET", fmt.Sprintf("%s/sshkeypairs", prefix), auth.Authenticate(sshKeysHandler), nil, "get_sshkeys", nil)
	app.AddHandler2("GET", fmt.Sprintf("%s/sshkeypairs/<tenant_id>", prefix), auth.Authenticate(adminSshKeysHandler), nil, "get_sshkeys", nil)
}

func adminSshKeysHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	publicOnly := false
	userCred := auth.FetchUserCredential(ctx)
	if !userCred.IsSystemAdmin() {
		publicOnly = true
	}
	params := appctx.AppContextParams(ctx)
	projectId := params["<tenant_id>"]
	if len(projectId) == 0 {
		httperrors.InputParameterError(w, "empty project_id/tenant_id")
		return
	}
	tenant, err := db.TenantCacheManager.FetchTenantByIdOrName(ctx, projectId)
	if err != nil {
		if err == sql.ErrNoRows {
			httperrors.ResourceNotFoundError(w, "tenant/project %s not found", projectId)
			return
		} else {
			httperrors.GeneralServerError(w, err)
			return
		}
	}
	query, err := jsonutils.ParseQueryString(r.URL.RawQuery)
	if err != nil {
		httperrors.GeneralServerError(w, err)
		return
	}
	isAdmin := jsonutils.QueryBoolean(query, "admin", false)

	sendSshKey(ctx, w, userCred, tenant.Id, isAdmin, publicOnly)
}

func sshKeysHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userCred := auth.FetchUserCredential(ctx)
	query, err := jsonutils.ParseQueryString(r.URL.RawQuery)
	if err != nil {
		httperrors.GeneralServerError(w, err)
		return
	}
	isAdmin := jsonutils.QueryBoolean(query, "admin", false)

	sendSshKey(ctx, w, userCred, userCred.GetProjectId(), isAdmin, false)
}

func sendSshKey(ctx context.Context, w http.ResponseWriter, userCred mcclient.TokenCredential, projectId string, isAdmin bool, publicOnly bool) {
	var privKey, pubKey string

	if isAdmin && userCred.IsSystemAdmin() {
		privKey, pubKey, _ = GetSshAdminKeypair(ctx)
	} else {
		privKey, pubKey, _ = GetSshProjectKeypair(ctx, projectId)
	}

	ret := jsonutils.NewDict()

	if !publicOnly {
		ret.Add(jsonutils.NewString(privKey), "private_key")
	}
	ret.Add(jsonutils.NewString(pubKey), "public_key")
	body := jsonutils.NewDict()
	body.Add(ret, "sshkeypair")
	appsrv.SendJSON(w, body)
}