package main

import (
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/models"
	_ "github.com/pbillerot/beerama/routers"
)

var err error

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
	if val, ok := config.String("thumbnail"); ok == nil {
		models.Config.Thumbnail = val
	}
	if val, ok := config.Int("width"); ok == nil {
		models.Config.Width = uint(val)
	}
	if val, ok := config.Int("height"); ok == nil {
		models.Config.Height = uint(val)
	}

	// lecture des répertoires dans beeDir
	loadBeedirs()
}

func loadBeedirs() {
	err = models.LoadBeeDirs()
	if err != nil {
		logs.Error("LoadBeeDirs", err)
	}
	// déclaration des répertoires racine et thumbnail en static
	web.SetStaticPath("/album", models.Config.Racine)
	web.SetStaticPath("/thumb", models.Config.Thumbnail)

}
