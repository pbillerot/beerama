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
	// user_id, _, ok := c.Ctx.Request.BasicAuth()
	// if !ok {
	// 	logs.Error("user_id inconnu")
	// } else {
	// 	c.Data["is_admin"] = models.Config.Users[user_id].IsAdmin
	// 	c.Data["is_editor"] = models.Config.Users[user_id].IsEditor
	// 	c.Data["is_reader"] = models.Config.Users[user_id].IsReader
	// }
	// c.Data["user_id"] = user_id

	if user_id, ok := c.GetSession("user_id").(string); ok {
		c.Data["user_id"] = user_id
		c.Data["is_admin"] = models.Config.Users[user_id].IsAdmin
		c.Data["is_editor"] = models.Config.Users[user_id].IsEditor
		c.Data["is_reader"] = models.Config.Users[user_id].IsReader
		if models.Config.Users[user_id].IsAdmin {
			c.Data["role"] = "admin"
		} else if models.Config.Users[user_id].IsEditor {
			c.Data["role"] = "editor"
		} else {
			c.Data["role"] = ""
		}
	} else {
		// pas de session ou expirée
		user_id, _, ok := c.Ctx.Request.BasicAuth()
		if !ok {
			logs.Error("user_id inconnu")
		} else {
			logs.Info("new session", user_id)
			c.Ctx.Output.Session("user_id", user_id)
			c.Data["user_id"] = user_id
			c.Data["is_admin"] = models.Config.Users[user_id].IsAdmin
			c.Data["is_editor"] = models.Config.Users[user_id].IsEditor
			c.Data["is_reader"] = models.Config.Users[user_id].IsReader
			if models.Config.Users[user_id].IsAdmin {
				c.Data["role"] = "admin"
			} else if models.Config.Users[user_id].IsEditor {
				c.Data["role"] = "editor"
			} else {
				c.Data["role"] = ""
			}
		}
	}
	c.Ctx.Output.Session("role", c.Data["role"].(string))

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
