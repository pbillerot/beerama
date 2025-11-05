package models

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/beego/beego/v2/core/logs"
	"github.com/oklog/ulid"
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
	fmt.Println("LoadBeeDirs. Proceeding.")
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
			text := strings.ReplaceAll(strings.TrimSpace(bfile.Description+" "+strings.Join(bfile.Keywords, " ")+" "+bfile.Make+" "+bfile.Model+" "+bfile.Base), "  ", " ")
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
	logs.Info("Images", "indexées")
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
			err, _ := beeDir.AddBeeFile(pi.Path, false)
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
	parent.Keywords = keywords

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
	parent.KeywordsAlbum = keyUniqueSorted
	parent.CountAlbum = count
	// maj des countAlbum et keywordsAlbum des sous-dossiers
	for _, bdir := range Config.BeeDirs {
		if bdir.ParentID == parent.ID {
			bdir.CountAlbum = count
			bdir.KeywordsAlbum = keyUniqueSorted
		}
	}

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
		parent := Config.BeeDirs[beeDir.ParentID]
		pathOld = Config.Racine + "/" + parent.Name + "/" + beeDir.Name
		pathNew = Config.Racine + "/" + parent.Name + "/" + newName
		originalOld = Config.Original + "/" + parent.Name + "/" + beeDir.Name
		originalNew = Config.Original + "/" + parent.Name + "/" + newName
		thumbOld = Config.Thumbnail + "/" + parent.Name + "/" + beeDir.Name
		thumbNew = Config.Thumbnail + "/" + parent.Name + "/" + newName
	}
	// le répertoire de l'image
	err := os.Rename(pathOld, pathNew)
	if err != nil {
		log.Fatalf("Failed to rename directory: %s en %s : %v", pathOld, pathNew, err)
		return err
	}
	// le répertoire de l'original
	_, err = os.Stat(originalOld)
	if os.IsExist(err) {
		err = os.Rename(originalOld, originalNew)
		if err != nil {
			log.Fatalf("Failed to rename directory: %s en %s : %v", originalOld, originalNew, err)
			return err
		}
	}
	// le répertoire de la vignette
	_, err = os.Stat(thumbOld)
	if os.IsExist(err) {
		err = os.Rename(thumbOld, thumbNew)
		if err != nil {
			log.Fatalf("Failed to rename directory: %s en %s : %v", thumbOld, thumbNew, err)
			return err
		}
	}
	// rename de beeDir
	beeDir.Name = newName
	beeDir.Path = pathNew
	return nil
}

// Rename beeFile.path en newName
func (beeFile *BeeFile) Rename(newName string) error {

	// rename du fichier, original et thumbnail
	var pathOld, pathNew, originalOld, originalNew, thumbOld, thumbNew string
	pathOld = beeFile.Path
	pathNew = strings.Replace(beeFile.Path, beeFile.Base, newName, 1)
	originalOld = beeFile.Original
	originalNew = strings.Replace(beeFile.Original, beeFile.Base, newName, 1)
	thumbOld = beeFile.Thumb
	thumbNew = strings.Replace(beeFile.Thumb, beeFile.Base, newName, 1)

	err := os.Rename(pathOld, pathNew)
	if err != nil {
		return err
	}
	if !beeFile.IsUrl {
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
		beeFile.UrlThumb = strings.Replace(beeFile.UrlThumb, beeFile.Base, newName, 1)
	}
	beeFile.UrlImage = strings.Replace(beeFile.UrlImage, beeFile.Base, newName, 1)
	return nil
}

// création id si à blanc 01K8ZHD17V12CV5SA29K1DW6TM
func (beeFile *BeeFile) GetNewId() string {
	if len(beeFile.ID) != 26 {
		// Use current time for the timestamp part
		t := time.Now().UTC()

		// Use crypto/rand or math/rand for the randomness part (entropy)
		entropy := rand.New(rand.NewSource(t.UnixNano()))

		serialNumber := ulid.MustNew(ulid.Timestamp(t), entropy)

		// e.g., 01AN4Z07BY79KA1307SR9X4MV3
		beeFile.ID = serialNumber.String()
		if beeFile.IsUrl {
			beeFile.UpdateFileUrl()
		} else {
			// report des meta dans l'image
			beeFile.UpdateMeta()
		}
		logs.Info("new beeid:", beeFile.ID, beeFile.Path)
	}
	return beeFile.ID
}

// Update du fichier.url
func (beeFile *BeeFile) UpdateFileUrl() error {

	fileUrl := FileUrl{}
	fileUrl.Description = beeFile.Description
	fileUrl.DateOriginal = beeFile.DateOriginal
	fileUrl.TimeOriginal = beeFile.TimeOriginal
	fileUrl.Keywords = beeFile.Keywords
	fileUrl.InternetShortcut.URL = beeFile.UrlImage
	fileUrl.Id = beeFile.GetNewId()

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
func (beeFile *BeeFile) Idx() {
	if Config.BeeFiles[beeFile.ID] == nil {
		Config.BeeFiles[beeFile.ID] = beeFile
	}
}
func GetBeeFile(id string) *BeeFile {
	if Config.BeeFiles[id] == nil {
		return &BeeFile{}
	} else {
		return Config.BeeFiles[id]
	}
}
