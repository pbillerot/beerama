package controllers

import (
	"fmt"
	"os"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	auth "github.com/beego/beego/v2/server/web/filter/auth"
	"github.com/pbillerot/beerama/models"
	"gopkg.in/yaml.v3"
)

// SecretAuth : Contrôle existence du username
// dans le fichier défini dans app.conf.users
func SecretAuth(username, password string) bool {
	if models.Config.Users[username].Password == password {
		return true
	} else {
		logs.Error("Connexion [%s]/[%s]", username, password)
	}
	return false
}

// DeclareAuth : Installation du filtre de contrôle d'accès à l'application
// voir dans main.go
func DeclareAuth() {
	yamlFile, err := os.ReadFile(models.Config.UsersFile)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", models.Config.UsersFile, err)
		logs.Error("Open", msg)
	}
	err = yaml.Unmarshal(yamlFile, &models.Config.Users)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", models.Config.UsersFile, err)
		logs.Error("Unmarshal", msg)
	}

	filter := auth.NewBasicAuthenticator(SecretAuth, "Contrôle d'accès")
	beego.InsertFilter("*", beego.BeforeRouter, filter)
}
