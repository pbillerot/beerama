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

// var err error

// MainController as
type MainController struct {
	beego.Controller
}

// Prepare implements Prepare method for loggedRouter.
func (c *MainController) Prepare() {
	// parametre de l'url

	// Contexte lié à app.conf
	c.Data["config"] = &models.Config

	// Initialisation des données de la session
	c.Data["sessionid"] = c.Ctx.GetCookie("beegosessionID")

	// folder en cours
	if param, ok := c.GetSession("folder").(string); ok {
		logs.Debug("folder:", param)
	} else {
		c.SetSession("folder", "/")
	}
	// Recherche en cours
	if param, ok := c.GetSession("search").(string); ok {
		c.Data["search"] = param
	} else {
		c.Data["search"] = ""
	}

	// admin or not admin
	if boolValue, ok := c.GetSession("is_admin").(bool); ok {
		if boolValue {
			c.Data["is_admin"] = true
		} else {
			// is admin ko
			c.Data["is_admin"] = false
		}
	} else {
		c.Data["is_admin"] = false
	}

	// XSRF protection des formulaires
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	// Sera ajouté derrière les urls pour ne pas utiliser le cache des images dynamiques
	c.Data["composter"] = time.Now().Unix()
	c.Data["refresh"] = false
}
