package models

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/beego/beego/v2/core/logs"
)

// Lecture du fichier beeusers.conf et chargement dans la structure
func LoadUsers() (err error) {

	content, err := os.ReadFile(Config.UsersPath)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("Open", msg)
		return err
	}
	err = toml.Unmarshal(content, &Config.Users)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("Unmarshal", msg)
		return err
	}
	logs.Info("LoadUsers. Proceeding.")
	return err
}

// Lecture du fichier beeusers.conf et chargement dans la structure
func GetUsersContent() (content string, err error) {

	buf, err := os.ReadFile(Config.UsersPath)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("ReadFile", msg)
		return
	}
	return string(buf), err
}

// Mise à jour du fichiers beeusers.conf et rechargement dans la structure
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

func (config *BeeConfig) IsUserAdmin(user_id string) bool {

	if access, ok := config.Users[user_id]; ok {
		return access.IsAdmin
	}

	return false
}

func (beeDir *BeeDir) IsUserEditor(user_id string) bool {

	if Config.IsUserAdmin(user_id) {
		return true
	}

	if beeDir.ParentID != "" {
		bdir := Config.BeeDirs[beeDir.ParentID]
		if access, ok := bdir.Users[user_id]; ok {
			return access.IsEditor
		}
	} else {
		if access, ok := beeDir.Users[user_id]; ok {
			return access.IsEditor
		}
	}
	return false
}

func (beeDir *BeeDir) IsUserReader(user_id string) bool {

	if Config.IsUserAdmin(user_id) {
		return true
	}

	if beeDir.ParentID != "" {
		bdir := Config.BeeDirs[beeDir.ParentID]
		if _, ok := bdir.Users[user_id]; ok {
			return true
		}
	} else {
		if _, ok := beeDir.Users[user_id]; ok {
			return true
		}
	}

	return false
}

// Mise à jour du fichiers .beeaccess.conf et rechargement dans la structure
func (beeDir *BeeDir) UpdateAccess(content []byte) (err error) {
	pathAccess := beeDir.Path + "/.beeaccess.conf"
	err = os.WriteFile(pathAccess, content, 0644)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", pathAccess, err)
		logs.Error("WriteFile", msg)
		return err
	}
	err = beeDir.LoadAccess()
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", Config.UsersPath, err)
		logs.Error("LoadAccess", msg)
		return err
	}

	return err
}

// Lecture du fichier beeusers.conf et chargement dans la structure
func (beeDir *BeeDir) GetAccessContent() (content string, err error) {
	pathAccess := beeDir.Path + "/.beeaccess.conf"
	buf, err := os.ReadFile(pathAccess)
	if err != nil {
		// lecture du modèle
		buf, _ := os.ReadFile("./conf/beeaccess-exemple.conf")
		return string(buf), nil
	}
	return string(buf), nil
}

// Lecture du fichier beeaccess.conf et chargement dans la structure beedir s'il existe
func (beeDir *BeeDir) LoadAccess() (err error) {
	pathAccess := beeDir.Path + "/.beeaccess.conf"

	buf, err := os.ReadFile(pathAccess)
	if err != nil {
		// n'existe pas ou erreur
		return nil
	}

	err = toml.Unmarshal(buf, &beeDir.Users)
	if err != nil {
		msg := fmt.Sprintf("%s : [%v]", pathAccess, err)
		logs.Error("Unmarshal", msg)
		return err
	}
	return nil
}
