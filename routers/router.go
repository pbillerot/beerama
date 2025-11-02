package routers

import (
	"fmt"

	"github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/controllers"
)

func init() {

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
		web.NSRouter("/image/:beefileid", &controllers.MainController{}, "get:Image"),
	)

	// Routes avec rôle reader
	nsReaderFolder := web.NewNamespace("/folder",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.ReaderRoleProfile),
		web.NSRouter("/:beedirid", &controllers.MainController{}, "get:Folder"),
		web.NSRouter("/download/:beedirid", &controllers.MainController{}, "post:FolderDownload"),
		web.NSRouter("/:beedirid/:htagid", &controllers.MainController{}, "get:FolderHtag"),
		web.NSRouter("/search/:beedirid", &controllers.MainController{}, "post:Search"),
	)
	nsReader := web.NewNamespace("/search",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.ReaderRoleProfile),
		web.NSRouter("/:beedirid", &controllers.MainController{}, "post:Search"),
	)
	// Routes avec rôle editor
	nsEditor := web.NewNamespace("/e",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Editor
		web.NSBefore(controllers.AuthRequiredProfile, controllers.EditorRoleProfile),
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
		web.NSRouter("/newurl/:beedirid", &controllers.MainController{}, "post:NewUrl"),
		web.NSRouter("/restore/:beedirid/:beefileid", &controllers.MainController{}, "post:Restore"),
	)

	// Routes avec rôle admin
	nsAdmin := web.NewNamespace("/a",
		// Enchaîne plusieurs filtres : l'utilisateur doit être authentifié ET Admin
		web.NSBefore(controllers.AuthRequiredProfile, controllers.AdminRoleProfile),
		web.NSRouter("/access/:beedirid", &controllers.MainController{}, "get:Access;post:Access"),
		web.NSRouter("/mkdir", &controllers.MainController{}, "post:MkFolder"),
		web.NSRouter("/reload", &controllers.MainController{}, "get:ReloadAll"),
		web.NSRouter("/users", &controllers.MainController{}, "get:Users;post:Users"),
	)

	// Ajouter les Namespaces au routeur Beego
	web.AddNamespace(nsStatic, nsReader, nsReaderFolder, nsEditor, nsAdmin)

	fmt.Println("Routes init. Proceeding.")
}
