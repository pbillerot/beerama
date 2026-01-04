package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	context "github.com/beego/beego/v2/server/web/context"
	"github.com/pbillerot/beerama/models"
)

// AuthRequiredProfile : Vérifie si l'utilisateur est authentifié
var AuthRequiredProfile = func(ctx *context.Context) {
	if models.Config.Runmode == "dev" {
		ctx.Request.Header.Add("Remote-User", models.Config.UserDev)
		ctx.Request.Header.Add("Remote-Groups", models.Config.GroupDev)
	}
	logs.Debug("Input:", fmt.Sprintf("%+v\n", ctx.Request.Header))
	user_id := ctx.Input.Header("Remote-User")
	if user_id == "" {
		logs.Error("BasiAuth not retrieve")
		ctx.Redirect(403, "/")
		return
	}
}

// EditorRoleProfile : Vérifie si l'utilisateur a le rôle d'éditeur de l'album courant
var EditorRoleProfile = func(ctx *context.Context) {

	user_id := ctx.Input.Header("Remote-User")
	groups := ctx.Input.Header("Remote-Groups")

	beeDir := models.Config.BeeDirs[ctx.Input.Param(":beedirid")]
	if beeDir.ParentID != "" {
		beeDir = models.Config.BeeDirs[beeDir.ParentID]
	}

	if !strings.Contains(groups, "admin") && !beeDir.IsUserEditor(user_id) {
		logs.Info("Filter: User is not Editor. Blocking.")
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		ctx.Output.Body([]byte("403 Forbidden - Editor required"))
	}
}

// AdminRoleProfile : Vérifie si l'utilisateur a le rôle d'administrateur
var AdminRoleProfile = func(ctx *context.Context) {
	groups := ctx.Input.Header("Remote-Groups")
	if !strings.Contains(groups, "admin") {
		logs.Info("Filter: User is not Admin. Blocking.")
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		ctx.Output.Body([]byte("403 Forbidden - Admin required"))
	}
}

// ReaderRoleProfile : Vérifie si l'utilisateur est authentifié
var ReaderRoleProfile = func(ctx *context.Context) {
	user_id := ctx.Input.Header("Remote-User")
	groups := ctx.Input.Header("Remote-Groups")

	beeDir := models.Config.BeeDirs[ctx.Input.Param(":beedirid")]
	if beeDir.ParentID != "" {
		beeDir = models.Config.BeeDirs[beeDir.ParentID]
	}

	if !strings.Contains(groups, "admin") && !beeDir.IsUserReader(user_id) {
		logs.Info("Filter: User is not Reader. Blocking.")
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		ctx.Output.Body([]byte("403 Forbidden - Reader required"))
		// panic("Stop")
	}
}
