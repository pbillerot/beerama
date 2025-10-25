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

	// Récupération du user_id contrôlé par beeauth
	// Retrieve the user_id and password from the *http.Request
	user_id, _, ok := c.Ctx.Request.BasicAuth()
	if !ok {
		logs.Error("BasiAuth not retrieve")
		return
	}

	c.Data["user_id"] = c.GetSession("user_id")
	c.Data["role"] = c.GetSession("role")
	c.Data["is_admin"] = models.Config.Users[user_id].IsAdmin

	// Contexte lié à app.conf
	c.Data["config"] = &models.Config

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
