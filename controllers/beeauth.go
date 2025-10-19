package controllers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/auth"
	"github.com/pbillerot/beerama/models"
)

// SecretAuth : Contrôle existence du user_id
// dans le fichier défini dans app.conf.users
func SecretAuth(user_id, password string) bool {
	return models.CheckUser(user_id, password)
}

// DeclareAuth : Installation du filtre de contrôle d'accès à l'application
// voir dans main.go
func DeclareAuth() {

	models.LoadUsers()

	filter := auth.NewBasicAuthenticator(SecretAuth, "Contrôle d'accès")
	web.InsertFilter("*", web.BeforeRouter, filter)
}
