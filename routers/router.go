package routers

import (
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	context "github.com/beego/beego/v2/server/web/context"
	"github.com/pbillerot/beerama/controllers"
	"github.com/pbillerot/beerama/models"
)

func init() {

	web.Router("/", &controllers.MainController{}, "get:Main")
	web.Router("/folder/:beedirid", &controllers.MainController{}, "get:Folder")
	web.Router("/return", &controllers.MainController{}, "get:Return")
	web.Router("/folder/:beedirid/:htagid", &controllers.MainController{}, "get:FolderHtag")
	web.Router("/search/:beedirid", &controllers.MainController{}, "post:Search")

	nsEditor := web.NewNamespace("/e", // editor
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(EditorRoleProfile),
		web.NSRouter("/rename/:beedirid", &controllers.MainController{}, "post:FolderRename"),
		web.NSRouter("/reload/:beedirid", &controllers.MainController{}, "get:Reload"),
		web.NSRouter("/meta/:beedirid/:beefileid", &controllers.MainController{}, "get:Meta;post:Meta"),
		web.NSRouter("/tag/:beedirid/:beefileid", &controllers.MainController{}, "post:Tag"),
		web.NSRouter("/upload/:beedirid", &controllers.MainController{}, "post:Upload"),
		web.NSRouter("/rm/:beedirid", &controllers.MainController{}, "post:FileRm"),
		web.NSRouter("/duplicate/:beedirid", &controllers.MainController{}, "post:Duplicate"),
		web.NSRouter("/cp/:beedirid", &controllers.MainController{}, "post:FileCopy"),
		web.NSRouter("/mv/:beedirid", &controllers.MainController{}, "post:FileMove"),
		web.NSRouter("/mkdir/:beedirid", &controllers.MainController{}, "post:MkSubFolder"),
		web.NSRouter("/rmdir/:beedirid", &controllers.MainController{}, "post:Rmdir"),
		web.NSRouter("/dragdrop/:beedirid", &controllers.MainController{}, "post:DragDrop"),
		web.NSRouter("/newdraw/:beedirid", &controllers.MainController{}, "post:NewDraw"),
		web.NSRouter("/restore/:beedirid/:beefileid", &controllers.MainController{}, "post:Restore"),
	)

	nsAdmin := web.NewNamespace("/a", // admin
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Admin
		web.NSBefore(AdminRoleProfile),
		web.NSRouter("/mkdir", &controllers.MainController{}, "post:MkFolder"),
		web.NSRouter("/users", &controllers.MainController{}, "get:Users;post:Users"),
		web.NSRouter("/reload", &controllers.MainController{}, "get:ReloadAll"),
	)

	// Ajouter les Namespaces au routeur Beego
	web.AddNamespace(nsEditor, nsAdmin)
}

// SecretAuth : Contrôle existence du user_id
// dans le fichier défini dans app.conf.users
func SecretAuth(user_id, password string) bool {
	return models.CheckUser(user_id, password)
}

// EditorRoleProfile : Vérifie si l'utilisateur a le rôle d'administrateur
var EditorRoleProfile = func(ctx *context.Context) {
	if ctx.Input.Session("role") != "editor" {
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
// var AuthRequiredProfile = func(ctx *context.Context) {
// 	if ctx.Input.Session("user_id") == nil {
// 		user_id, _, ok := ctx.Request.BasicAuth()
// 		if !ok {
// 			auth.NewBasicAuthenticator(SecretAuth, "Auth check")
// 		} else {
// 			ctx.Output.Session("user_id", user_id)
// 			if models.Config.Users[user_id].IsAdmin {
// 				ctx.Output.Session("role", "admin")
// 			} else if models.Config.Users[user_id].IsEditor {
// 				ctx.Output.Session("role", "editor")
// 			} else {
// 				ctx.Output.Session("role", "user")
// 			}
// 		}

// 		// fmt.Println("Filter: User not authenticated. Blocking.")
// 		// ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
// 		// ctx.Output.Body([]byte("401 Unauthorized"))

// 		// Arrêter le traitement de la requête
// 		// panic("Stop")
// 	}
// 	// fmt.Println("Filter: User authenticated. Proceeding.")
// }
