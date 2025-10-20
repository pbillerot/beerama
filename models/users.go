package models

import (
	"fmt"
	"os"

	"github.com/beego/beego/v2/core/logs"
	"gopkg.in/yaml.v3"
)

// Lecture du fichier users.yaml et chargement dans la structure
func LoadUsers() (err error) {

	yamlFile, err := os.ReadFile(Config.UsersPath)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("Open", msg)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &Config.Users)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("Unmarshal", msg)
		return err
	}
	fmt.Println("LoadUsers. Proceeding.")
	return err
}

// Lecture du fichier users.yaml et chargement dans la structure
func GetUsersContent() (content string, err error) {

	buf, err := os.ReadFile(Config.UsersPath)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("ReadFile", msg)
		return
	}
	return string(buf), err
}

// Mise à jour du fichiers users.yaml et rechargement dans la structure
func UpdateUsers(content []byte) (err error) {

	err = os.WriteFile(Config.UsersPath, content, 0644)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("WriteFile", msg)
		return err
	}
	err = LoadUsers()
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("LoadUsers", msg)
		return err
	}

	return err
}

func CheckUser(user_id, password string) bool {
	if Config.Users[user_id].Password == password {
		return true
	} else {
		logs.Error("Tentative Connexion [%s]/[%s]", user_id, password)
	}
	return false
}
