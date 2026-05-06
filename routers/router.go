package routers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"beerama/controllers"
)

func init() {

	// Filter 1: BeforeRouter - Record the request start time
	web.InsertFilter("*", web.BeforeRouter, controllers.BeforeRouter)

	// Filter 2: FinishRouter - Log the access details after the request is processed
	web.InsertFilter("*", web.FinishRouter, controllers.LogAccessFilter, web.WithReturnOnOutput(false))

	// Authentification obligatoire pour toutes les routes
	// filter := auth.NewBasicAuthenticator(controllers.SecretAuth, "Basic Authentification")
	// web.InsertFilter("*", web.BeforeRouter, filter)
	web.InsertFilter("*", web.BeforeExec, controllers.BasicAuthFilter)

	// Routes sans rôle particulier
	web.Router("/", &controllers.MainController{}, "get:Main")
	web.Router("/return", &controllers.MainController{}, "get:Return")

	// Routes static pour protéger les photos des albums
	nsStatic := web.NewNamespace("/s",
		web.NSBefore(controllers.AuthRequiredProfile),
		web.NSRouter("/document/:beefileid", &controllers.MainController{}, "get:Document"),
		web.NSRouter("/doc/:beefileid", &controllers.MainController{}, "get:Doc"),
	)

	// Routes avec rôle reader
	nsReaderFolder := web.NewNamespace("/folder",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.ReaderRoleProfile),
		web.NSRouter("/:beedirid", &controllers.MainController{}, "get:Folder"),
		web.NSRouter("/download/:beedirid", &controllers.MainController{}, "post:Download"),
		web.NSRouter("/:beedirid/:htag", &controllers.MainController{}, "get:HTag"),
	)
	nsReader := web.NewNamespace("/search",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.ReaderRoleProfile),
		web.NSRouter("/:beedirid", &controllers.MainController{}, "get:Search;post:Search"),
	)
	// Routes avec rôle editor
	nsEditor := web.NewNamespace("/e",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.EditorRoleProfile),
		web.NSRouter("/rename/:beedirid", &controllers.MainController{}, "post:FolderRename"),
		web.NSRouter("/reload/:beedirid", &controllers.MainController{}, "get:Reload"),
		web.NSRouter("/meta/:beedirid/:beefileid", &controllers.MainController{}, "get:Meta;post:Meta"),
		web.NSRouter("/lot/:beedirid", &controllers.MainController{}, "post:Lot"),
		web.NSRouter("/tag/:beedirid/:beefileid", &controllers.MainController{}, "post:Tag"),
		web.NSRouter("/upload/:beedirid", &controllers.MainController{}, "post:Upload"),
		web.NSRouter("/rm/:beedirid", &controllers.MainController{}, "post:FileRm"),
		web.NSRouter("/duplicate/:beedirid", &controllers.MainController{}, "post:Duplicate"),
		web.NSRouter("/cp/:beedirid", &controllers.MainController{}, "post:FileCopy"),
		web.NSRouter("/mv/:beedirid", &controllers.MainController{}, "post:FileMove"),
		web.NSRouter("/mkdir/:beedirid", &controllers.MainController{}, "post:MkSubFolder"),
		web.NSRouter("/rmdir/:beedirid", &controllers.MainController{}, "post:Rmdir"),
		web.NSRouter("/dragdrop/:beedirid", &controllers.MainController{}, "post:DragDrop"),
		web.NSRouter("/newdoc/:beedirid", &controllers.MainController{}, "post:NewDoc"),
		web.NSRouter("/newdraw/:beedirid", &controllers.MainController{}, "post:NewDraw"),
		web.NSRouter("/newurl/:beedirid", &controllers.MainController{}, "post:NewUrl"),
		web.NSRouter("/restore/:beedirid/:beefileid", &controllers.MainController{}, "post:Restore"),
	)

	// Routes avec rôle admin
	nsAdmin := web.NewNamespace("/a",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Admin
		web.NSBefore(controllers.AuthRequiredProfile, controllers.AdminRoleProfile),
		web.NSRouter("/access/:beedirid", &controllers.MainController{}, "get:Access;post:Access"),
		web.NSRouter("/rmdir/:beedirid", &controllers.MainController{}, "post:Rmdir"),
		web.NSRouter("/mkdir", &controllers.MainController{}, "post:MkFolder"),
		web.NSRouter("/reload", &controllers.MainController{}, "get:ReloadAll"),
		web.NSRouter("/users", &controllers.MainController{}, "get:Users;post:Users"),
	)

	// Ajouter les Namespaces au routeur Beego
	web.AddNamespace(nsStatic, nsReader, nsReaderFolder, nsEditor, nsAdmin)

	logs.Info("Routes init. Proceeding.")
}
