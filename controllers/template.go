package controllers

import (
	"slices"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

//
//	Fonctions pour les templates
//

// Déclaration des fonctions utilisées dans les templates
func init() {
	web.AddFuncMap("BeeSubstr", BeeSubstr)
	web.AddFuncMap("BeeIN", BeeIN)
	web.AddFuncMap("BeeReplace", BeeReplace)
	web.AddFuncMap("BeeSplit", BeeSplit)
	web.AddFuncMap("BeeToString", BeeToString)
	web.AddFuncMap("BeeLower", BeeLower)
	web.AddFuncMap("BeeContains", BeeContains)
	logs.Info("Template.init. Proceeding.")
}

// beeLower
func BeeLower(s string) (out string) {
	return strings.ToLower(s)
}

// beeSubstr
func BeeSubstr(in string, start, end int) (out string) {
	if start > len(in) || end > len(in) {
		return in
	}
	return in[start:end]
}

// BeeToString as
func BeeToString(list []string) (out string) {
	return strings.Join(list, " ")
}

// BeeContains as
func BeeContains(buf string, in string) bool {
	return strings.Contains(buf, in)
}

// BeeIN as
func BeeIN(list []string, in string) bool {
	if in == "" {
		return true
	}
	return slices.Contains(list, in)
}

// BeeSplit strings séparées par un séparateur en slice
func BeeSplit(in string, separateur string) (out []string) {
	if in != "" {
		out = strings.Split(in, separateur)
	} else {
		out = []string{}
	}
	return
}

// BeeReplace as
func BeeReplace(in string, old string, new string) (out string) {
	out = strings.Replace(in, old, new, 1)
	return
}
