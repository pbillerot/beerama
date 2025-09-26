package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/controllers"
)

func init() {

	web.Router("/", &controllers.MainController{}, "get:Main")
	web.Router("/folder/:beedirid", &controllers.MainController{}, "get:Folder")
	web.Router("/rename/:beedirid", &controllers.MainController{}, "post:FolderRename")
	web.Router("/reload/", &controllers.MainController{}, "get:ReloadAll")
	web.Router("/reload/:beedirid", &controllers.MainController{}, "get:Reload")
	web.Router("/folder/:beedirid/:htagid", &controllers.MainController{}, "get:FolderHtag")
	web.Router("/meta/:beedirid/:beefileid", &controllers.MainController{}, "get:Meta;post:Meta")
	web.Router("/tag/:beedirid/:beefileid", &controllers.MainController{}, "post:Tag")
	web.Router("/restore/:beedirid/:beefileid", &controllers.MainController{}, "post:Restore")
	web.Router("/upload/:beedirid", &controllers.MainController{}, "post:Upload")
	web.Router("/rm/:beedirid", &controllers.MainController{}, "post:FileRm")
	web.Router("/duplicate/:beedirid", &controllers.MainController{}, "post:Duplicate")
	web.Router("/cp/:beedirid", &controllers.MainController{}, "post:FileCopy")
	web.Router("/mv/:beedirid", &controllers.MainController{}, "post:FileMove")
	web.Router("/mkdir", &controllers.MainController{}, "post:MkFolder")
	web.Router("/mkdir/:beedirid", &controllers.MainController{}, "post:MkSubFolder")
	web.Router("/rmdir/:beedirid", &controllers.MainController{}, "post:Rmdir")
	web.Router("/dragdrop/:beedirid", &controllers.MainController{}, "post:DragDrop")

	web.Router("/admin", &controllers.MainController{}, "get:Admin")

}
