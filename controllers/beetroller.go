package controllers

/**
	MainController
	Gestion de la session
**/
import (
	"fmt"
	"html/template"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	context "github.com/beego/beego/v2/server/web/context"
	"github.com/pbillerot/beerama/models"
)

// logAccessFilter is the core logging function registered at FinishRouter
var BeforeRouter = func(ctx *context.Context) {
	ctx.Input.SetData("startTime", time.Now())
}

// logAccessFilter is the core logging function registered at FinishRouter
var LogAccessFilter = func(ctx *context.Context) {

	startTime, ok := ctx.Input.GetData("startTime").(time.Time)
	if !ok {
		// Fallback if start time wasn't set for some reason
		startTime = time.Now()
	}

	// 2. Calculate details
	duration := time.Since(startTime)
	clientIP := ctx.Input.IP() // Recommended way to get the client IP in Beego
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	// Get the response status code
	statusCode := ctx.ResponseWriter.Status
	if statusCode == 0 {
		statusCode = 200 // Default to 200 if no explicit status was set
	}

	// 3. Define your custom log format
	// Example Format: [IP] METHOD PATH STATUS DURATION
	logMessage := fmt.Sprintf("[%s] %s %s %d %v",
		clientIP,
		method,
		path,
		statusCode,
		duration,
	)

	// 4. Write the log message using Beego's logger
	// Use logs.Notice or logs.Info depending on your log level configuration
	logs.Notice(logMessage)
}

// MainController as
type MainController struct {
	beego.Controller
}

// Prepare implements Prepare method for loggedRouter.
func (c *MainController) Prepare() {

	user_id := c.Ctx.Input.Header("Remote-User")

	c.Data["user_id"] = user_id
	c.Data["is_admin"] = models.Config.Users[user_id].IsAdmin

	// Contexte lié à app.conf
	c.Data["config"] = &models.Config

	// folder en cours
	if _, ok := c.GetSession("folder").(string); ok {
		// logs.Trace("folder:", param)
	} else {
		c.SetSession("folder", "/")
	}
	// Recherche en cours
	if param, ok := c.GetSession("search").(string); ok {
		c.Data["search"] = param
	} else {
		c.Data["search"] = ""
	}

	// mémorisation de l'adresse IP
	// clientIP := c.Ctx.Input.IP()

	// XSRF protection des formulaires
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	// Sera ajouté derrière les urls pour ne pas utiliser le cache des images dynamiques
	c.Data["composter"] = time.Now().Unix()
	c.Data["refresh"] = false
}

func GlobalFormatter(lm *logs.LogMsg) string {
	return fmt.Sprintf("[GLOBAL] %s", lm.Msg)
}
