package controllers

/**
	MainController
	Gestion de la session
**/
import (
	"html/template"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/models"
)

// MainController as
type MainController struct {
	beego.Controller
}

// Prepare implements Prepare method for loggedRouter.
func (c *MainController) Prepare() {

	// Récupération du username contrôlé par beeauth
	// Retrieve the username and password from the *http.Request
	username, _, ok := c.Ctx.Request.BasicAuth()
	if !ok {
		logs.Error("Username inconnu")
	} else {
		c.Data["is_admin"] = models.Config.Users[username].IsAdmin
		c.Data["is_editor"] = models.Config.Users[username].IsEditor
		c.Data["is_reader"] = models.Config.Users[username].IsReader
	}
	c.Data["username"] = username

	// Contexte lié à app.conf
	c.Data["config"] = &models.Config

	// Initialisation des données de la session
	c.Data["sessionid"] = c.Ctx.GetCookie("beegosessionID")

	// folder en cours
	if _, ok := c.GetSession("folder").(string); ok {
		// logs.Debug("folder:", param)
	} else {
		c.SetSession("folder", "/")
	}
	// Recherche en cours
	if param, ok := c.GetSession("search").(string); ok {
		c.Data["search"] = param
	} else {
		c.Data["search"] = ""
	}

	// XSRF protection des formulaires
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	// Sera ajouté derrière les urls pour ne pas utiliser le cache des images dynamiques
	c.Data["composter"] = time.Now().Unix()
	c.Data["refresh"] = false
}
