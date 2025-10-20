package controllers

import (
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
	context "github.com/beego/beego/v2/server/web/context"
	"github.com/pbillerot/beerama/models"
)

// SecretAuth : Contrôle existence du user_id
// dans le fichier défini dans app.conf.users
func SecretAuth(user_id, password string) bool {
	return models.CheckUser(user_id, password)
}

// EditorRoleProfile : Vérifie si l'utilisateur a le rôle d'administrateur
var EditorRoleProfile = func(ctx *context.Context) {
	if ctx.Input.Session("role") != "admin" && ctx.Input.Session("role") != "editor" {
		fmt.Println("Filter: User is not Editor. Blocking.")
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		ctx.Output.Body([]byte("403 Forbidden - Editor required"))
		// panic("Stop")
	}
	// fmt.Println("Filter: User is Editor. Proceeding.")
}

// AdminRoleProfile : Vérifie si l'utilisateur a le rôle d'administrateur
var AdminRoleProfile = func(ctx *context.Context) {
	if ctx.Input.Session("role") != "admin" {
		fmt.Println("Filter: User is not Admin. Blocking.")
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		ctx.Output.Body([]byte("403 Forbidden - Admin required"))
		// panic("Stop")
	}
	// fmt.Println("Filter: User is Admin. Proceeding.")
}

// AuthRequiredProfile : Vérifie si l'utilisateur est authentifié
var AuthRequiredProfile = func(ctx *context.Context) {
	user_id, _, ok := ctx.Request.BasicAuth()
	if !ok {
		logs.Error("BasiAuth not retrieve")
		return
	}
	if ctx.Input.Session("user_id") == nil {
		ctx.Output.Session("user_id", user_id)
		if models.Config.Users[user_id].IsAdmin {
			ctx.Output.Session("role", "admin")
		} else if models.Config.Users[user_id].IsEditor {
			ctx.Output.Session("role", "editor")
		} else {
			ctx.Output.Session("role", "user")
		}
		logs.Info("new session", user_id)
	}
	// fmt.Println("Filter: User authenticated. Proceeding.")
}
