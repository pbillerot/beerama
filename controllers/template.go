package controllers

import (
	"slices"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/models"
)

//
//	Fonctions pour les templates
//

// Déclaration des fonctions utilisées dans les templates
func init() {
	beego.AddFuncMap("BeeIN", BeeIN)
	beego.AddFuncMap("BeeSplitBreadcrumb", BeeSplitBreadcrumb)
	beego.AddFuncMap("BeeReplace", BeeReplace)
	beego.AddFuncMap("BeeSplit", BeeSplit)
	beego.AddFuncMap("BeeToString", BeeToString)
}

// BeeToString as
func BeeToString(list []string) (out string) {
	return strings.Join(list, " ")
}

// BeeIN as
func BeeIN(list []string, in string) bool {
	if in == "" {
		return true
	}
	return slices.Contains(list, in)
}

// BeeSplitBreadcrumb /rep1/rep2/rep3/file.ext
func BeeSplitBreadcrumb(path string) (breadcrumb []models.Breadcrumb) {
	reps := strings.Split(path, "/")
	pp := ""
	ll := len(reps)
	isLast := false
	for i, rep := range reps {
		if i == 0 {
			continue
		}
		pp = pp + "/" + rep
		if i == (ll - 1) {
			isLast = true
		}
		breadcrumb = append(breadcrumb, models.Breadcrumb{Base: rep, Path: pp, IsLast: isLast})
	}
	return
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
