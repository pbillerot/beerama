package main

import (
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"beerama/models"

	// pour charger les init() des packages suivants
	_ "beerama/controllers"
	_ "beerama/routers"
)

func main() {
	web.Run()
}

func init() {
	// Initialisation de models.Config
	if val, ok := config.String("github"); ok == nil {
		models.Config.Github = val
	}
	if val, ok := config.String("help"); ok == nil {
		models.Config.Help = val
	}
	if val, ok := config.String("version"); ok == nil {
		models.Config.Version = val
		logs.Info("version", val)
	}
	if val, ok := config.String("appname"); ok == nil {
		models.Config.AppName = val
	}
	if val, ok := config.String("title"); ok == nil {
		models.Config.Title = val
	}
	if val, ok := config.String("description"); ok == nil {
		models.Config.Description = val
	}
	if val, ok := config.String("favicon"); ok == nil {
		models.Config.Favicon = val
	}
	if val, ok := config.String("background"); ok == nil {
		models.Config.Background = val
	}
	if val, ok := config.String("icon"); ok == nil {
		models.Config.Icon = val
	}
	if val, ok := config.String("racine"); ok == nil {
		models.Config.Racine = val
	}
	if val, ok := config.String("original"); ok == nil {
		models.Config.Original = val
	}
	if val, ok := config.String("trash"); ok == nil {
		models.Config.Trash = val
	}
	if val, ok := config.String("thumbnail"); ok == nil {
		models.Config.Thumbnail = val
	}
	if val, ok := config.String("users"); ok == nil {
		models.Config.UsersPath = val
	}
	if val, ok := config.String("index"); ok == nil {
		models.Config.IndexDirs = val
	}
	if val, ok := config.Int("width"); ok == nil {
		models.Config.Width = int(val)
	}
	if val, ok := config.Int("height"); ok == nil {
		models.Config.Height = int(val)
	}
	if val, ok := config.Bool("debug"); ok == nil {
		if val {
			logs.SetLevel(logs.LevelDebug)
		} else {
			logs.SetLevel(logs.LevelInfo)
		}
	}

	logs.Critical("1 critical")
	logs.Error("2 error")
	logs.Warning("3 warning")
	logs.Trace("4 debug")
	logs.Info("5 info")
	logs.Trace("6 trace")
	// chargement des users
	models.LoadUsers()

	// lecture des répertoires dans beeDir
	err := models.LoadBeeDirs()
	if err != nil {
		logs.Error("LoadBeeDirs", err)
	}

	// déclaration des répertoires racine et thumbnail en static et protégé par login
	web.SetStaticPath("/s/album", models.Config.Racine)
	web.SetStaticPath("/s/thumb", models.Config.Thumbnail)

	logs.Info("Main init. Proceeding.")
}
