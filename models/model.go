package models

import (
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/beego/beego/v2/core/logs"
)

// Config de config.yaml lu dans main.init()
var Config BeeConfig

// LoadBeeDirs chargement de la liste des répertoires BeeDir
func LoadBeeDirs() error {
	var pis []BeePathInfo
	// dossiers racines
	err := getOnlyFolders(Config.Racine, &pis)
	if err != nil {
		return err
	}
	// Instanciation des BeeDirs
	Config.BeeDirs = Config.BeeDirs[:0]
	var dirid = 0
	for _, pi := range pis {
		var beeDir BeeDir
		beeDir.ID = "dir" + strconv.Itoa(dirid)
		beeDir.ParentID = ""
		beeDir.Path = pi.Path
		beeDir.Name = pi.Info.Name()
		err := beeDir.LoadBeeFiles(0)
		if err != nil {
			return err
		}
		Config.BeeDirs = append(Config.BeeDirs, &beeDir)
		dirid = dirid + 1
	}
	sort.Slice(Config.BeeDirs, func(i, j int) bool {
		return Config.BeeDirs[i].Name < Config.BeeDirs[j].Name
	})

	// sous-dossiers
	for _, beeDir := range Config.BeeDirs {
		pis = pis[:0]
		err = getOnlyFolders(Config.Racine+"/"+beeDir.Name, &pis)
		if err != nil {
			return err
		}
		// ajout des images du sous-dossier
		for _, pi := range pis {
			var bdir BeeDir
			bdir.ID = "dir" + strconv.Itoa(dirid)
			bdir.ParentID = beeDir.ID
			bdir.Path = pi.Path
			bdir.Name = pi.Info.Name()
			err := bdir.LoadBeeFiles(len(beeDir.BeeFiles) + 1)
			if err != nil {
				return err
			}
			beeDir.WithChildren = true
			Config.BeeDirs = append(Config.BeeDirs, &bdir)
			dirid = dirid + 1
			// beeDir.BeeFiles = append(beeDir.BeeFiles, bdir.BeeFiles...)

		}
		beeDir.UpdateBeeDir()
	}

	return err
}

// AddFolder
func (config *BeeConfig) AddFolder(path string) {
	var beeDir BeeDir
	beeDir.ID = "dir" + strconv.Itoa(len(config.BeeDirs))
	beeDir.Path = config.Racine + "/" + path
	beeDir.Name = path
	config.BeeDirs = append(config.BeeDirs, &beeDir)
	sort.Slice(config.BeeDirs, func(i, j int) bool {
		return config.BeeDirs[i].Name < config.BeeDirs[j].Name
	})
}

// AddSubFolder
func (config *BeeConfig) AddSubFolder(parent *BeeDir, name string) {
	parent.WithChildren = true
	var beedir BeeDir
	beedir.ID = "dir" + strconv.Itoa(len(config.BeeDirs))
	beedir.Path = config.Racine + "/" + parent.Name + "/" + name
	beedir.Name = name
	beedir.ParentID = parent.ID
	config.BeeDirs = append(config.BeeDirs, &beedir)
	sort.Slice(config.BeeDirs, func(i, j int) bool {
		return config.BeeDirs[i].Name < config.BeeDirs[j].Name
	})
}

// RemoveFolder
func (config *BeeConfig) RemoveFolder(beeDir *BeeDir) {

	err := os.RemoveAll(beeDir.Path)
	if err != nil {
		return
	}
	// suppression du beeDir de conig.BeeDirs
	// recherche de l'indice dans le tableau
	for index, bdir := range config.BeeDirs {
		if bdir.Path == beeDir.Path {
			config.BeeDirs = append(config.BeeDirs[:index], config.BeeDirs[index+1:]...)
			break
		}
	}
	sort.Slice(config.BeeDirs, func(i, j int) bool {
		return config.BeeDirs[i].Name < config.BeeDirs[j].Name
	})
}

// readFolder retourne la liste des fichiers dans BeePathInfo
func getOnlyFolders(directory string, info *[]BeePathInfo) (err error) {
	// ouverture du dossier
	f, err := os.Open(directory)
	if err != nil {
		return
	}
	// lecture des fichiers et dossiers du dossier courant
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return
	}
	// tri des dossiers sur le nom inversé si numérique
	sort.Slice(list, func(i, j int) bool {
		if _, err := strconv.Atoi(list[i].Name()); err == nil {
			if _, err := strconv.Atoi(list[j].Name()); err == nil {
				return list[i].Name() > list[j].Name()
			}
			return list[i].Name() < list[j].Name()
		}
		return list[i].Name() < list[j].Name()
	})
	// tri des fichiers sur le nom
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	// Rangement des dossiers visibles (non préfixés par un .)
	for _, file := range list {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			var pi BeePathInfo
			pi.Path = directory + "/" + file.Name()
			pi.Info = file
			*info = append(*info, pi)
		}
	}
	return
}

// LoadBeeFiles chargement des fichier de BeeDir
func (beeDir *BeeDir) LoadBeeFiles(idstart int) error {
	// Exiftool
	et, err := exiftool.NewExiftool()
	if err != nil {
		return err
	}
	defer et.Close()

	// lecture du répertoire
	var pis []BeePathInfo
	err = readFolder(beeDir.Path, &pis)
	if err != nil {
		return err
	}

	beeDir.BeeFiles = beeDir.BeeFiles[:0]
	for _, pi := range pis {
		if !pi.Info.IsDir() {
			err, _ := beeDir.AddBeeFile(pi.Path, idstart)
			if err != nil {
				continue
			}
		}
	}
	// maj beedir
	beeDir.UpdateBeeDir()
	logs.Info(beeDir.Name, "rechargé")
	return nil
}

// UpdateBeeDir - maj keywords tri - tri des beefiles sur la date original
// suite ajout suppression d'une image et modification date et keyword
func (beeDir *BeeDir) UpdateBeeDir() {
	// ajout des keywords doublon et tri
	beeDir.Keywords = beeDir.Keywords[:0]
	// les keywords de l'album
	for _, beeFile := range beeDir.BeeFiles {
		beeDir.Keywords = append(beeDir.Keywords, beeFile.Keywords...)
		// beeFile.ID = "id" + strconv.Itoa(id)
	}
	// ajout des keywords des sous-dossiers
	for _, bdir := range Config.BeeDirs {
		if bdir.ParentID == beeDir.ID {
			beeDir.Keywords = append(beeDir.Keywords, bdir.Keywords...)
		}
	}
	keyUniqueSorted := BeeUniqueString(beeDir.Keywords)
	sort.Strings(keyUniqueSorted)
	beeDir.Keywords = keyUniqueSorted
	// tri des images sur la date original
	sort.Slice(beeDir.BeeFiles, func(i, j int) bool {
		return beeDir.BeeFiles[i].DateOriginal < beeDir.BeeFiles[j].DateOriginal
	})
}

// RenameBeeDir - beeDir.Name Path Dir Original Thumb UrlImage UrlThumb
// - repertoire album et sous-dossiers, originals et thumbs
func (beeDir *BeeDir) RenameBeeDir(newName string) error {
	// rename des répertoire album, original et thumbnail
	var pathOld, pathNew, originalOld, originalNew, thumbOld, thumbNew string
	if beeDir.ParentID == "" {
		pathOld = Config.Racine + "/" + beeDir.Name
		pathNew = Config.Racine + "/" + newName
		originalOld = Config.Original + "/" + beeDir.Name
		originalNew = Config.Original + "/" + newName
		thumbOld = Config.Thumbnail + "/" + beeDir.Name
		thumbNew = Config.Thumbnail + "/" + newName
	} else {
		parent := GetBeeDir(beeDir.ParentID)
		pathOld = Config.Racine + "/" + parent.Name + "/" + beeDir.Name
		pathNew = Config.Racine + "/" + parent.Name + "/" + newName
		originalOld = Config.Original + "/" + parent.Name + "/" + beeDir.Name
		originalNew = Config.Original + "/" + parent.Name + "/" + newName
		thumbOld = Config.Thumbnail + "/" + parent.Name + "/" + beeDir.Name
		thumbNew = Config.Thumbnail + "/" + parent.Name + "/" + newName
	}
	err := os.Rename(pathOld, pathNew)
	if err != nil {
		return err
	}
	_, err = os.Stat(originalOld)
	if os.IsExist(err) {
		err = os.Rename(originalOld, originalNew)
		if err != nil {
			return err
		}
	}
	err = os.Rename(thumbOld, thumbNew)
	if err != nil {
		return err
	}
	// // rename des chemins des fichiers de l'album dans beeFile
	// for _, beeFile := range beeDir.BeeFiles {
	// 	if beeDir.ParentID == "" {
	// 		beeFile.Dir = Config.Racine + "/" + newName
	// 		beeFile.Path = beeFile.Dir + "/" + beeFile.Base
	// 		beeFile.Original = Config.Original + "/" + newName + "/" + beeFile.Base
	// 		beeFile.Thumb = Config.Thumbnail + "/" + newName + "/th_" + beeFile.Base
	// 		beeFile.UrlImage = "/album/" + newName + "/" + beeFile.Base
	// 		beeFile.UrlThumb = "/thumb/" + newName + "/th_" + beeFile.Base
	// 	} else {
	// 		beeFile.Dir = Config.Racine + "/" + beeDir.Name + "/" + newName
	// 		beeFile.Path = beeFile.Dir + "/" + beeDir.Name + "/" + beeFile.Base
	// 		beeFile.Original = Config.Original + "/" + beeDir.Name + "/" + newName + "/" + beeFile.Base
	// 		beeFile.Thumb = Config.Thumbnail + "/" + beeDir.Name + "/" + newName + "/th_" + beeFile.Base
	// 		beeFile.UrlImage = "/album" + "/" + beeDir.Name + "/" + newName + "/" + beeFile.Base
	// 		beeFile.UrlThumb = "/thumb" + "/" + beeDir.Name + "/" + newName + "/th_" + beeFile.Base
	// 	}
	// }
	// if beeDir.WithChildren {
	// 	for _, bdir := range Config.BeeDirs {
	// 		if bdir.ParentID == beeDir.ID {
	// 			for _, beeFile := range bdir.BeeFiles {
	// 				beeFile.Dir = Config.Racine + "/" + newName + "/" + bdir.Name
	// 				beeFile.Path = beeFile.Dir + "/" + beeFile.Base
	// 				beeFile.Original = Config.Original + "/" + newName + "/" + bdir.Name + "/" + beeFile.Base
	// 				beeFile.Thumb = Config.Thumbnail + "/" + newName + "/" + bdir.Name + "/th_" + beeFile.Base
	// 				beeFile.UrlImage = "/album/" + newName + "/" + bdir.Name + "/" + beeFile.Base
	// 				beeFile.UrlThumb = "/thumb/" + newName + "/" + bdir.Name + "/th" + beeFile.Base
	// 			}
	// 		}
	// 	}
	// }
	// rename de beeDir
	beeDir.Name = newName
	beeDir.Path = pathNew
	return nil
}

// AddKeywords ajout dans beedir, suppression des doublons, tri des clés
func (beeDir *BeeDir) AddKeywords(keywords []string) {
	beeDir.Keywords = append(beeDir.Keywords, keywords...)
	beeDir.UpdateBeeDir()
}

// AddKeyword ajout dans beedir.Keywordw, suppression des doublons, tri des clés
func (beeDir *BeeDir) AddKeyword(keyword string) {
	beeDir.Keywords = append(beeDir.Keywords, keyword)
	beeDir.UpdateBeeDir()
	if beeDir.ParentID != "" {
		// report des keywords dans l'album
		parent := GetBeeDir(beeDir.ParentID)
		parent.Keywords = append(parent.Keywords, beeDir.Keywords...)
		parent.UpdateBeeDir()
	}
}

func BeeUniqueString(s []string) []string {
	// Crée une map pour stocker les éléments uniques.
	keys := make(map[string]bool)
	// Crée un slice pour le résultat final.
	list := []string{}

	// Parcours le slice d'entrée.
	for _, entry := range s {
		// Vérifie si la clé (chaîne) existe déjà dans la map.
		if _, value := keys[entry]; !value {
			// Si la clé n'existe pas, ajoute-la à la map et au slice de résultat.
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// GetBeePathDir retourne la BeeDir
func GetBeePathDir(path string) *BeeDir {
	for _, dir := range Config.BeeDirs {
		if path == dir.Path {
			return dir
		}
	}
	return &BeeDir{}
}

// GetBeeDir retourne la BeeDir
func GetBeeDir(beedirid string) *BeeDir {
	for _, dir := range Config.BeeDirs {
		if beedirid == dir.ID {
			return dir
		}
	}
	return &BeeDir{}
}

// GetFirstBeeDir retourne la première BeeDir
func GetFirstBeeDir() *BeeDir {
	for _, dir := range Config.BeeDirs {
		return dir
	}
	return &BeeDir{}
}

// GetBeeDir retourne la BeeDir
func GetBeeFile(beedirid, beefileid string) *BeeFile {
	for _, dir := range Config.BeeDirs {
		if beedirid == dir.ID {
			for _, file := range dir.BeeFiles {
				if beefileid == file.ID {
					return file
				}
			}
		}
	}
	return &BeeFile{}
}

// GetBeeDir retourne la BeeDir qui correspond au path
func GetBeeFilePath(beeDir *BeeDir, path string) *BeeFile {
	for _, file := range beeDir.BeeFiles {
		if path == file.Path {
			return file
		}
	}
	return &BeeFile{}
}

// readFolder retourne la liste des fichiers dans BeePathInfo
func readFolder(dirname string, info *[]BeePathInfo) (err error) {
	// ouverture du dossier
	f, err := os.Open(dirname)
	if err != nil {
		return
	}
	// lecture ds fichiers et dossiers du dossier courant
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return
	}
	// tri des dossiers sur le nom inversé si numérique
	sort.Slice(list, func(i, j int) bool {
		if _, err := strconv.Atoi(list[i].Name()); err == nil {
			if _, err := strconv.Atoi(list[j].Name()); err == nil {
				return list[i].Name() > list[j].Name()
			}
			return list[i].Name() < list[j].Name()
		}
		return list[i].Name() < list[j].Name()
	})
	// // tri des fichiers sur le nom
	// sort.Slice(list, func(i, j int) bool {
	// 	return list[i].Name() < list[j].Name()
	// })
	// Rangement des dossiers au début
	for _, file := range list {
		if file.IsDir() {
			var pi BeePathInfo
			pi.Path = dirname + "/" + file.Name()
			pi.Info = file
			*info = append(*info, pi)
		}
	}
	// Rangement des fichiers à la fin
	for _, file := range list {
		if !file.IsDir() {
			var pi BeePathInfo
			pi.Path = dirname + "/" + file.Name()
			pi.Info = file
			*info = append(*info, pi)
		}
	}
	return
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
