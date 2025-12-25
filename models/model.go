package models

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/beego/beego/v2/core/logs"
	"github.com/pbillerot/beerama/fulltext"
	"github.com/pbillerot/beerama/shutil"
)

// Config de config.yaml lu dans main.init()
var Config BeeConfig

// LoadBeeDirs chargement de la liste des répertoires BeeDir
func LoadBeeDirs() error {
	Config.BeeFiles = make(map[string]*BeeFile)
	var pis []BeePathInfo
	// dossiers racines
	err := getOnlyFolders(Config.Racine, &pis)
	if err != nil {
		logs.Error(err)
		return err
	}
	// Instanciation des BeeDirs
	Config.BeeDirs = make(map[string]*BeeDir)
	var dirid = 0
	for _, pi := range pis {
		var beeDir BeeDir
		beeDir.ID = "dir" + strconv.Itoa(dirid)
		beeDir.ParentID = ""
		beeDir.Path = pi.Path
		beeDir.Name = pi.Info.Name()
		err := beeDir.LoadBeeFiles()
		if err != nil {
			logs.Error(err)
			return err
		}
		// chargement des beeaccess.conf
		beeDir.LoadAccess()

		Config.BeeDirs[beeDir.ID] = &beeDir
		dirid = dirid + 1
	}
	// sous-dossiers
	var childrens = make(map[string]*BeeDir)
	for _, beeDir := range Config.BeeDirs {
		var pis []BeePathInfo
		err = getOnlyFolders(Config.Racine+"/"+beeDir.Name, &pis)
		if err != nil {
			logs.Error(err)
			return err
		}
		// ajout des images du sous-dossier
		for _, pi := range pis {
			var bdir BeeDir
			bdir.ID = "dir" + strconv.Itoa(dirid)
			bdir.ParentID = beeDir.ID
			bdir.Path = pi.Path
			bdir.Name = pi.Info.Name()
			err := bdir.LoadBeeFiles()
			if err != nil {
				logs.Error(err)
				return err
			}
			childrens[bdir.ID] = &bdir
			dirid = dirid + 1
		}
	}
	// on ajoute les enfants
	for _, bdir := range childrens {
		Config.BeeDirs[bdir.ID] = bdir
	}
	// Calcul des compteurs et keywords des dossiers
	for _, bdir := range Config.BeeDirs {
		bdir.UpdateAlbum()
	}
	err = Config.IndexAllBeefiles()
	logs.Info("LoadBeeDirs. Proceeding.")
	return err
}

// IndexAllBeefiles pour tous les albums
func (config *BeeConfig) IndexAllBeefiles() error {

	// create new index with temp dir (usually "" is fine)
	idx, err := fulltext.NewIndexer(string(config.IndexDirs))
	if err != nil {
		logs.Error(err)
		return err
	}
	defer idx.Close()

	// indexation par le moteur fulltext intégré dans les sources
	// https://github.com/bradleypeabody/fulltext
	for _, bdir := range config.BeeDirs {

		for _, bfile := range bdir.BeeFiles {
			// provide stop words if desired
			idx.StopWordCheck = fulltext.FrenchStopWordChecker

			// for each document you want to add, you do something like this:
			text := strings.ReplaceAll(strings.TrimSpace(bfile.Description+" "+strings.Join(bfile.Keywords, " ")+" "+bfile.Make+" "+bfile.Model+" "+bfile.Title+" "+bfile.ID), "  ", " ")
			doc := fulltext.IndexDoc{
				Id:         []byte(bdir.ID + "_" + bfile.ID), // unique identifier (the path to a webpage works...)
				StoreValue: []byte(text),                     // bytes you want to be able to retrieve from search results
				IndexValue: []byte(config.Index),             // bytes you want to be split into words and indexed
			}
			// logs.Info(string(doc.Id), string(doc.IndexValue), string(doc.StoreValue))
			idx.AddDoc(doc) // add it
		}

	}
	// when done, write out to final index
	f, err := os.Create(config.IndexDirs + "/idxout")
	if err != nil {
		logs.Error(err)
		return err
	}

	err = idx.FinalizeAndWrite(f)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Trace("Images", "indexées")
	return err
}

// AddFolder
func (config *BeeConfig) AddFolder(path string) {
	var beeDir BeeDir
	beeDir.ID = "dir" + strconv.Itoa(len(config.BeeDirs))
	beeDir.Path = config.Racine + "/" + path
	beeDir.Name = path
	beeDir.BeeFiles = make(map[string]*BeeFile)
	config.BeeDirs[beeDir.ID] = &beeDir
}

// AddSubFolder
func (config *BeeConfig) AddSubFolder(parent *BeeDir, name string) {
	parent.WithChildren = true
	var beeDir BeeDir
	beeDir.ID = "dir" + strconv.Itoa(len(config.BeeDirs))
	beeDir.Path = config.Racine + "/" + parent.Name + "/" + name
	beeDir.Name = name
	beeDir.ParentID = parent.ID
	beeDir.BeeFiles = make(map[string]*BeeFile)
	config.BeeDirs[beeDir.ID] = &beeDir
	// config.BeeDirs = append(config.BeeDirs, &beedir)
	// sort.Slice(config.BeeDirs, func(i, j int) bool {
	// 	return config.BeeDirs[i].Name < config.BeeDirs[j].Name
	// })
}

// RemoveFolder
func (config *BeeConfig) RemoveFolder(beeDir *BeeDir) {

	err := os.RemoveAll(beeDir.Path)
	if err != nil {
		logs.Error(err)
		return
	}
	// suppression du beeDir de conig.BeeDirs
	// recherche de l'indice dans le tableau
	delete(config.BeeDirs, beeDir.ID)
	// for index, bdir := range config.BeeDirs {
	// 	if bdir.Path == beeDir.Path {
	// 		config.BeeDirs = append(config.BeeDirs[:index], config.BeeDirs[index+1:]...)
	// 		break
	// 	}
	// }
	// sort.Slice(config.BeeDirs, func(i, j int) bool {
	// 	return config.BeeDirs[i].Name < config.BeeDirs[j].Name
	// })
}

// readFolder retourne la liste des fichiers dans BeePathInfo
func getOnlyFolders(directory string, info *[]BeePathInfo) (err error) {
	var pis []BeePathInfo
	err = readFolder(directory, &pis)
	if err != nil {
		logs.Error(err)
		return err
	}
	for _, pi := range pis {
		if strings.HasPrefix(filepath.Base(pi.Path), ".") {
			continue
		}
		if pi.Info.IsDir() {
			*info = append(*info, pi)
		}
		if shutil.IsSymlink(pi.Info) {
			*info = append(*info, pi)
		}
	}
	// tri des fichiers sur le nom
	sort.Slice(pis, func(i, j int) bool {
		return pis[i].Info.Name() < pis[j].Info.Name()
	})
	return
}

// LoadBeeFiles chargement des fichier de BeeDir
func (beeDir *BeeDir) LoadBeeFiles() error {

	// lecture du répertoire
	var pis []BeePathInfo
	err := readFolder(beeDir.Path, &pis)
	if err != nil {
		logs.Error(err)
		return err
	}
	beeDir.BeeFiles = make(map[string]*BeeFile) // beeDir.BeeFiles[:0]
	for _, pi := range pis {
		if strings.Contains(pi.Path, "/.") {
			continue
		}
		if !pi.Info.IsDir() {
			err, _ := beeDir.CreateBeeFile(pi.Path, false)
			if err != nil {
				continue
			}
		}
	}
	// tri des images sur la date et l'heure original
	keys := make([]string, 0, len(beeDir.BeeFiles))
	for k := range beeDir.BeeFiles {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if beeDir.BeeFiles[keys[i]].DateOriginal != beeDir.BeeFiles[keys[j]].DateOriginal {
			return beeDir.BeeFiles[keys[i]].DateOriginal < beeDir.BeeFiles[keys[j]].DateOriginal
		}
		return beeDir.BeeFiles[keys[i]].TimeOriginal < beeDir.BeeFiles[keys[j]].TimeOriginal
	})
	if beeDir.ParentID == "" {
		logs.Info("Album:", beeDir.Name, "rechargé")
	} else {
		logs.Info("...", beeDir.Name, "rechargé")
	}

	return nil
}

// CreateBeeFile
func (beeDir *BeeDir) CreateBeeFile(path string, isNew bool) (*BeeFile, error) {
	beeFile := &BeeFile{}

	// calcul des chemins pour calculer le type de fichier ci-après
	beeFile.ComputePathsId(path)
	beeFile.DirID = beeDir.ID

	if Contains([]string{".jpeg", ".jpg", ".png"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
		if strings.Contains(beeFile.Base, ".drawio.") {
			beeFile.IsDrawio = true
		}
		if strings.Contains(beeFile.Base, ".doc.") {
			beeFile.IsDoc = true
		}
	} else if Contains([]string{".gif"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
	} else if Contains([]string{".svg"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
		beeFile.IsSvg = true
	} else if Contains([]string{".mov", ".m4v", ".mkv", ".mp4", ".webm"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsVideo = true
	} else if Contains([]string{".pdf"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsPdf = true
	} else if Contains([]string{".conf"}, beeFile.Ext) {
		var content []byte
		content, err := os.ReadFile(beeFile.Path)
		if err != nil {
			logs.Error(err)
		}
		beeFile.Content = content
		beeFile.IsConf = true
	} else {
		err := fmt.Errorf("Extension fichier inconnue: %s", beeFile.Path)
		logs.Error(err)
		return beeFile, err
	}

	if beeFile.IsImage || beeFile.IsPdf || beeFile.IsVideo {
		beeFile.GetMetadata()
	}
	// title avec le nom du fichier aseptisé
	if beeFile.Title == "" {
		beeFile.Title = strings.Join(EclaterNomDeFichierEnMots(beeFile.Path), " ")
	}

	// calcul de la clé du fichier dans ID
	beeFile.GetNewId()
	// recalcul des chemins en fonctions du type de fichier et de l'ID qui à remplacé le name
	beeFile.ComputePathsId(path)

	// report des keywords dand beeDir
	if isNew {
		beeFile.Keywords = append(beeFile.Keywords, "new")
	}
	beeDir.Keywords = append(beeDir.Keywords, beeFile.Keywords...)

	if isNew {
		// suppression des étiquettes en doublons
		keyUniqueSorted := BeeUniqueString(beeFile.Keywords)
		sort.Strings(keyUniqueSorted)
		beeFile.Keywords = beeFile.Keywords[:0]
		beeFile.Keywords = append(beeFile.Keywords, keyUniqueSorted...)
	}

	// Renommage du fichier avec le ID
	if beeFile.ID != beeFile.Name {
		logs.Info("Renommage du fichier en %s de %s", beeFile.ID+beeFile.Ext, beeFile.Path)
		err := beeFile.Rename(beeFile.ID + beeFile.Ext)
		if err == nil {
			// titre en particulier
			beeFile.UpdateMeta()
		}
	}

	// création de la miniature dans Config.Thumbnail si n'existe pas
	// et si <> pdf doc
	if !beeFile.existeThumbnail() {
		beeFile.createThumbnail(Config.Width, Config.Height)
	}

	// ajout dans BeeFiles
	beeDir.BeeFiles[beeFile.ID] = beeFile

	// Indexation du beefile
	beeFile.Idx()

	return beeFile, nil
}

func (beeDir *BeeDir) GetParent() *BeeDir {
	var parent *BeeDir
	if beeDir.ParentID == "" {
		parent = beeDir
	} else {
		parent = GetBeeDir(beeDir.ParentID)
	}
	return parent
}

// Retourne les beedirs liés beedir courant accessibles par le user_id
// calcul du nombre de fichiers de l'album
func (beeDir *BeeDir) GetParentBeedirs() *[]BeeDir {
	var beedirs []BeeDir
	// sélection des albums accessibles par user_id liés à l'album courant parent
	beedirs = append(beedirs, *beeDir)
	// ajout des enfants de l'album
	children := []BeeDir{}
	for _, bdir := range Config.BeeDirs {
		if bdir.ParentID == beeDir.ID {
			children = append(children, *bdir)
		}
	}
	// tri des enfants
	sort.Slice(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})
	// ajout des enfants à la fin
	beedirs = append(beedirs, children...)

	return &beedirs
}

// UpdateBeeDir sans relire le répertoire
// calcul de count et keywords puis tri des beefiles
func (beeDir *BeeDir) UpdateBeeDir() {
	// calcul des compeurs keywords et tri des beefiles
	keywords := []string{}
	// les keywords de l'album
	count := 0
	for _, beeFile := range beeDir.BeeFiles {
		keywords = append(keywords, beeFile.Keywords...)
		count++
	}
	beeDir.Count = count
	// suppression des keywords en doublon
	keyUniqueSorted := BeeUniqueString(keywords)
	sort.Strings(keyUniqueSorted)
	beeDir.Keywords = keyUniqueSorted

	// tri des images sur la date et l'heure original
	keys := make([]string, 0, len(beeDir.BeeFiles))
	for k := range beeDir.BeeFiles {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if beeDir.BeeFiles[keys[i]].DateOriginal != beeDir.BeeFiles[keys[j]].DateOriginal {
			return beeDir.BeeFiles[keys[i]].DateOriginal < beeDir.BeeFiles[keys[j]].DateOriginal
		}
		return beeDir.BeeFiles[keys[i]].TimeOriginal < beeDir.BeeFiles[keys[j]].TimeOriginal
	})
}

// UpdateAlbum sans relire le répertoire
// pour tous les fichiers de l'album et sous-dossiers
// lance UpdateBeeDir
// compte le nombre de fichiers
// fusionne les keywords
func (beeDir *BeeDir) UpdateAlbum() {
	parent := beeDir.GetParent()

	// les keywords et compteur du dossier de l'album
	keywords := []string{}
	count := 0
	for _, beeFile := range parent.BeeFiles {
		keywords = append(keywords, beeFile.Keywords...)
		count++
	}
	// maj du dossier
	parent.Count = count
	// suppression des keywords en doublon
	keyUniqueSorted := BeeUniqueString(keywords)
	sort.Strings(keyUniqueSorted)
	parent.Keywords = parent.Keywords[:0]
	parent.Keywords = append(parent.Keywords, keyUniqueSorted...)

	// calcul des countAlbum et keywordsAlbum des sous-dossiers
	for _, bdir := range Config.BeeDirs {
		if bdir.ParentID == parent.ID {
			parent.WithChildren = true
			bdir.UpdateBeeDir()
			keywords = append(keywords, bdir.Keywords...)
			count += bdir.Count
		}
	}
	// suppression des keywords en doublon
	keyUniqueSorted = BeeUniqueString(keywords)
	sort.Strings(keyUniqueSorted)
	parent.KeywordsAlbum = parent.KeywordsAlbum[:0]
	parent.KeywordsAlbum = append(parent.KeywordsAlbum, keyUniqueSorted...)
	parent.CountAlbum = count
	// maj des countAlbum et keywordsAlbum des sous-dossiers
	for _, bdir := range Config.BeeDirs {
		if bdir.ParentID == parent.ID {
			bdir.CountAlbum = count
			bdir.KeywordsAlbum = bdir.KeywordsAlbum[:0]
			bdir.KeywordsAlbum = append(bdir.KeywordsAlbum, keyUniqueSorted...)
		}
	}

}

// RenameBeeDir - beeDir.Name Path Dir Original Thumb UrlImage UrlThumb
func (beeDir *BeeDir) RenameBeeDir(newName string) error {
	// renommage du dossier de l'album et des sous-dossiers
	for _, bDir := range Config.BeeDirs {
		var pathOld, pathNew, originalOld, originalNew, thumbOld, thumbNew string
		if bDir.ID == beeDir.ID {
			if beeDir.ParentID == "" {
				// album à renommer
				pathOld = Config.Racine + "/" + bDir.Name
				pathNew = Config.Racine + "/" + newName
				originalOld = Config.Original + "/" + bDir.Name
				originalNew = Config.Original + "/" + newName
				thumbOld = Config.Thumbnail + "/" + bDir.Name
				thumbNew = Config.Thumbnail + "/" + newName
			} else {
				// sous-dossier à renommer seulement
				parent := Config.BeeDirs[bDir.ParentID]
				pathOld = Config.Racine + "/" + parent.Name + "/" + bDir.Name
				pathNew = Config.Racine + "/" + parent.Name + "/" + newName
				originalOld = Config.Original + "/" + parent.Name + "/" + bDir.Name
				originalNew = Config.Original + "/" + parent.Name + "/" + newName
				thumbOld = Config.Thumbnail + "/" + parent.Name + "/" + bDir.Name
				thumbNew = Config.Thumbnail + "/" + parent.Name + "/" + newName
			}
			// le répertoire du dossier
			logs.Info("Rename directory: %s to %s", pathOld, pathNew)
			err := os.Rename(pathOld, pathNew)
			if err != nil {
				logs.Error("Failed to rename directory: %s en %s : %v", pathOld, pathNew, err)
				return err
			}
			// le répertoire de l'original
			_, err = os.Stat(originalOld)
			if !os.IsNotExist(err) {
				logs.Info("Rename directory: %s to %s", originalOld, originalNew)
				err = os.Rename(originalOld, originalNew)
				if err != nil {
					logs.Error("Failed to rename directory: %s en %s : %v", originalOld, originalNew, err)
					return err
				}
			}
			// le répertoire de la vignette
			_, err = os.Stat(thumbOld)
			if !os.IsNotExist(err) {
				logs.Info("Rename directory: %s to %s", thumbOld, thumbNew)
				err = os.Rename(thumbOld, thumbNew)
				if err != nil {
					logs.Error("Failed to rename directory: %s en %s : %v", thumbOld, thumbNew, err)
					return err
				}
			}
			// rename de beeDir
			bDir.Name = newName
			bDir.Path = pathNew
			// Rechargement des beefiles par relecture du répertoire
			bDir.LoadBeeFiles()
		} else if bDir.ParentID == beeDir.ID {
			// seul le répertoire du parent à changé
			parent := Config.BeeDirs[bDir.ParentID]
			pathNew = Config.Racine + "/" + parent.Name + "/" + bDir.Name
			bDir.Path = pathNew
			// Rechargement des beefiles par relecture du répertoire
			bDir.LoadBeeFiles()
		}
	}
	return nil
}

// Rename beeFile.path en newName
func (beeFile *BeeFile) Rename(newName string) error {

	// rename du fichier, original et thumbnail
	pathOld := beeFile.Path

	pathNew := strings.Replace(beeFile.Path, beeFile.Base, newName, 1)
	beeFile.ComputePathsId(pathNew)

	err := os.Rename(pathOld, pathNew)
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil
}

// Calcul des chemins du fichier
func (beeFile *BeeFile) ComputePathsId(path string) {
	beeFile.Path = path
	beeFile.Dir = filepath.Dir(path)
	beeFile.Base = filepath.Base(path)
	beeFile.Ext = filepath.Ext(path)
	beeFile.Name = strings.TrimSuffix(beeFile.Base, beeFile.Ext)
	dirOriginal := Config.Original + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	beeFile.Original = dirOriginal + beeFile.Base
	// Name dans le bon format ?
	if VerifierFormat(beeFile.Name) {
		beeFile.ID = beeFile.Name
	}
	beeFile.UrlImage = "/s/album" + beeFile.Dir[len(Config.Racine):] + "/" + beeFile.Base
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	if beeFile.IsPdf || beeFile.IsVideo {
		beeFile.Thumb = dirThumb + "th_" + beeFile.Base + ".jpg"
		beeFile.UrlThumb = "/s/thumb" + dirThumb[len(Config.Thumbnail):] + "th_" + beeFile.Base + ".jpg"
	} else {
		beeFile.Thumb = dirThumb + "th_" + beeFile.Base
		beeFile.UrlThumb = "/s/thumb" + dirThumb[len(Config.Thumbnail):] + "th_" + beeFile.Base
	}

}

// Update du fichier.url
func (beeFile *BeeFile) UpdateFileUrl() error {

	fileUrl := FileUrl{}
	fileUrl.Id = beeFile.ID
	fileUrl.Title = beeFile.Title
	fileUrl.Description = beeFile.Description
	fileUrl.DateOriginal = beeFile.DateOriginal
	fileUrl.TimeOriginal = beeFile.TimeOriginal
	fileUrl.Keywords = beeFile.Keywords
	fileUrl.InternetShortcut.URL = beeFile.UrlImage

	updatedData, err := toml.Marshal(&fileUrl)
	if err != nil {
		return err
	}
	if err := os.WriteFile(beeFile.Path, updatedData, 0644); err != nil {
		return err
	}
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
	if Config.BeeDirs[beedirid] == nil {
		return &BeeDir{}
	}
	return Config.BeeDirs[beedirid]
}

// GetFirstBeeDir retourne la première BeeDir
func GetFirstBeeDir() *BeeDir {
	for _, dir := range Config.BeeDirs {
		return dir
	}
	return &BeeDir{}
}

// GetBeeFilePath retourne la BeeDir qui correspond au path
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

// Contains checks if a string is present in a slice
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// ajout du beefile dans l'index
func (beeFile *BeeFile) Idx() {
	// if Config.BeeFiles[beeFile.ID] == nil {
	// 	Config.BeeFiles[beeFile.ID] = beeFile
	// }
	Config.BeeFiles[beeFile.ID] = beeFile
}
func GetBeeFile(id string) *BeeFile {
	if Config.BeeFiles[id] == nil {
		return &BeeFile{}
	} else {
		return Config.BeeFiles[id]
	}
}

// création id si à blanc aa-00-00-00
func (beeFile *BeeFile) GetNewId() string {
	if !VerifierFormat(beeFile.Name) {
		// key plus simple aa-00-00-00
		beeFile.ID = GenerateKey()
		if strings.Contains(beeFile.Name, ".drawio") {
			beeFile.ID += ".drawio"
		}
		// beeFile.Name = beeFile.ID
		logs.Info("new beeid:", beeFile.ID, beeFile.Path)
	} else {
		// est-ce que beefile existe déjà
		// if GetBeeFile(beeFile.Name).Name == beeFile.Name {
		// 	// il existe
		// 	// - genération d'une nouvelle clé pour éviter les doublons lors des uploads
		// 	beeFile.ID = GenerateKey()
		// 	logs.Info("new beeid:", beeFile.ID, beeFile.Path)
		// } else {
		// 	beeFile.ID = beeFile.Name
		// }
		beeFile.ID = beeFile.Name
	}
	return beeFile.ID
}

// eclaterNomDeFichierEnMots prend un chemin de fichier complet et le divise en mots,
// en gérant les séparateurs courants (-, _, .) et le CamelCase.
func EclaterNomDeFichierEnMots(cheminComplet string) []string {
	// 1. Isoler le nom du fichier sans chemin ni extension
	base := filepath.Base(cheminComplet)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	// 2. Remplacer les séparateurs courants par un espace unique
	baseNameAvecEspaces := baseName
	for _, sep := range []string{"_", "-", "."} {
		baseNameAvecEspaces = strings.ReplaceAll(baseNameAvecEspaces, sep, " ")
	}

	// 3. Gérer le CamelCase en insérant un espace avant chaque majuscule
	var result strings.Builder
	for i, r := range baseNameAvecEspaces {
		// Vérifie si le caractère actuel est une majuscule et n'est pas au début
		if unicode.IsUpper(r) && i > 0 {
			// Vérifie si le caractère précédent n'est pas un espace
			// Cela évite d'ajouter des espaces avant les majuscules qui suivent déjà un séparateur converti en espace.
			prevRune := rune(baseNameAvecEspaces[i-1])
			if !unicode.IsSpace(prevRune) {
				// Ajoute un espace seulement si le caractère précédent n'est pas déjà une majuscule (pour les acronymes comme 'HTML')
				if !unicode.IsUpper(prevRune) || (i+1 < len(baseNameAvecEspaces) && unicode.IsLower(rune(baseNameAvecEspaces[i+1]))) {
					result.WriteRune(' ')
				}
			}
		}
		result.WriteRune(r)
	}

	// 4. Éclater le résultat par les espaces blancs pour obtenir les mots finaux
	// strings.Fields supprime les espaces multiples et les espaces en début/fin.
	mots := strings.Fields(result.String())

	// Optionnel : convertir tous les mots en minuscules pour la cohérence
	for i, word := range mots {
		mots[i] = strings.ToLower(word)
	}

	return mots
}

// generateKey génère une clé de 11 caractères au format XX-00-00-00
func GenerateKey() string {
	// Initialiser le générateur aléatoire
	// rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(1))

	// 1. Générer les deux caractères alphanumériques (XX)
	const letterBytes = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 2)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	prefixe := string(b)

	// 2. Générer les trois groupes de deux chiffres (00-00-00)
	// On utilise rand.Intn(100) pour obtenir un nombre entre 0 et 99 inclus.
	n1 := rand.Intn(100)
	n2 := rand.Intn(100)
	n3 := rand.Intn(100)

	// 3. Formatter la clé
	// Le format "%02d" garantit que le nombre est affiché sur 2 chiffres
	// en le préfixant par un zéro si nécessaire (ex: 5 devient 05).
	key := fmt.Sprintf("%s-%02d-%02d-%02d", prefixe, n1, n2, n3)

	return key
}

// Expression régulière compilée du nom de fichier sans extension une fois pour toute
var RegNameCompiled = regexp.MustCompile(`^[a-zA-Z]{2}-\d{2}-\d{2}-\d{2}.*`)

// vérifie que la chaîne est du format xx-99-99-99
func VerifierFormat(chaine string) bool {
	// Compile l'expression régulière une seule fois.
	// Pour un usage intensif, il est préférable de compiler la regex une fois au démarrage
	// du programme et de réutiliser l'objet *regexp.Regexp.

	// Utilise MatchString pour vérifier si la chaîne correspond au modèle.
	return RegNameCompiled.MatchString(chaine)
}

// Expression régulière compilée de l'url openstreetmap une fois pour toute
var RegOSMCompiled = regexp.MustCompile(`#map=\d+/([^/]+)/([^/&]+)`)

func GetLatitudeLongitude(urlExterne string) (latitude, longitude string) {
	matches := RegOSMCompiled.FindStringSubmatch(urlExterne)
	if len(matches) >= 3 {
		// Le premier élément (index 0) est la correspondance complète.
		// L'index 1 est le premier groupe de capture (p1).
		p1 := matches[1]
		// L'index 2 est le deuxième groupe de capture (p2).
		p2 := matches[2]
		return p1, p2
	} else {
		logs.Error("Latitude longitude non trouveés dans %s", urlExterne)
	}
	return "48.866669", "2.33333" // mairie de Paris
}
