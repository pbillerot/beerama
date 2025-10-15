package controllers

/**
	MainController
	Gestion de la session
**/
import (
	"html/template"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
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

// The Filter function runs before the router executes the controller
var AuthFilter = func(ctx *beecontext.Context) {
	// 1. Get the Authorization header
	authHeader := ctx.Input.Header("Authorization")

	// Check if the header exists and starts with "Bearer "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		// Stop execution and send an Unauthorized response
		// ctx.Abort(401, "Unauthorized: Token missing or invalid format")
		return
	}

	// 2. Extract the token (e.g., stripping "Bearer ")
	tokenString := strings.Split(authHeader, " ")[1]

	// 3. Validate the token (You need a JWT library for this)
	// For demonstration, let's assume a successful validation for now.
	userID, isValid := validateAndExtractUserID(tokenString) // <--- Implement this function!

	if !isValid {
		ctx.Abort(401, "Unauthorized: Invalid token")
		return
	}

	// 4. Store the authenticated user ID in the context for later use in controllers
	ctx.Input.SetData("UserID", userID)

	// If successful, the request proceeds to the matched controller
}

// NOTE: This is a placeholder function; you must implement actual JWT logic.
func validateAndExtractUserID(token string) (int, bool) {
	// ... actual JWT parsing, verification, and claims extraction logic ...
	// For example:
	// claims, err := jwt.Parse(token, ...)
	// if err != nil || !claims.Valid { return 0, false }
	// return claims["user_id"].(int), true
	logs.Info(token)
	return 123, true // Placeholder for successful authentication
}
