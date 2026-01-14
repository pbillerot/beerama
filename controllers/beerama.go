package controllers

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/fulltext"
	"github.com/pbillerot/beerama/models"
	"github.com/pbillerot/beerama/shutil"
)

// Main as get and Post /
func (c *MainController) Main() {

	beego.ReadFromRequest(&c.Controller)

	// Sélection des bdirs accessibles par le user_id
	user_id := c.GetSession("user_id").(string)
	var beeDirs []models.BeeDir
	for _, bdir := range models.Config.BeeDirs {
		if bdir.ParentID == "" && bdir.IsUserReader(user_id) {
			beeDirs = append(beeDirs, *bdir)
		}
	}
	// tri des albums
	sort.Slice(beeDirs, func(i, j int) bool {
		return beeDirs[i].Name < beeDirs[j].Name
	})

	c.Data["beedirs"] = &beeDirs

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "index.html"
}

// Return à l'url mémorisée dans la session "folder" /return
func (c *MainController) Return() {
	// Retour sur url
	if url, ok := c.GetSession("folder").(string); ok {
		c.Ctx.Redirect(302, url)
	} else {
		c.Ctx.Redirect(302, "/folder")
	}
}

// Folder Sélection d'un folder /folder/:beedirid
func (c *MainController) Folder() {

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	if len(c.Data["search"].(string)) > 0 {
		c.Ctx.Redirect(302, "/search/"+beeDir.ID)
	}

	user_id := c.GetSession("user_id").(string)

	parent := beeDir.GetParent()

	// Sélection des bdirs des sous-dossiers du parent courant
	beeDirs := parent.GetParentBeedirs() // liste des albums autorisés pour un déplacement éventuel
	// liste des albums du user
	var beeAlbums []models.BeeDir
	for _, bdir := range models.Config.BeeDirs {
		if bdir.ParentID == "" && bdir.IsUserEditor(user_id) {
			beeAlbums = append(beeAlbums, *bdir)
		}
	}
	// tri des albums
	sort.Slice(beeAlbums, func(i, j int) bool {
		return beeAlbums[i].Name < beeAlbums[j].Name
	})

	// Construction de la liste des beefiles
	beeFiles := []models.BeeFile{}
	for _, bfile := range beeDir.BeeFiles {
		beeFiles = append(beeFiles, *bfile)
	}
	// tri des beefiles
	sort.Slice(beeFiles, func(i, j int) bool {
		var dateI = beeFiles[i].DateOriginal
		var dateJ = beeFiles[j].DateOriginal
		if beeFiles[i].Year != "" {
			dateI = beeFiles[i].Year + ":01:01"
		}
		if beeFiles[j].Year != "" {
			dateJ = beeFiles[j].Year + ":01:01"
		}
		return dateI < dateJ
	})
	c.Data["albums"] = &beeAlbums
	c.Data["parent"] = &parent
	c.Data["beedirs"] = &beeDirs
	c.Data["beedir"] = &beeDir
	c.Data["beefiles"] = &beeFiles
	c.Data["htagid"] = ""
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	// Mémorisation du dernier appel
	c.SetSession("folder", c.Ctx.Request.RequestURI)

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "folder.html"
}

// Download Téléchargement des fichiers sélectionnés
func (c *MainController) Download() {

	files := c.GetStrings("files[]")
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	if len(files) == 1 {
		beeFile := beeDir.BeeFiles[files[0]]
		c.Ctx.Output.Download(beeFile.Path, beeFile.Base)
		return
	}

	filesToZip := []string{}
	for _, beefile := range beeDir.BeeFiles {
		if slices.Contains(files, beefile.ID) {
			filesToZip = append(filesToZip, beefile.Path)
		}
	}

	// 1. Set the necessary headers for a ZIP file download
	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/zip")
	// The Content-Disposition header forces a download and suggests a filename.
	filename := fmt.Sprintf("beerama-%d.zip", time.Now().Unix())
	c.Ctx.ResponseWriter.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	// 2. Create a new zip writer that writes directly to the HTTP response body
	zipWriter := zip.NewWriter(c.Ctx.ResponseWriter)
	defer zipWriter.Close() // Ensure the zip writer is closed to finalize the archive

	// 3. Loop through the files and add them to the zip archive
	for _, filename := range filesToZip {
		// Open the file from the disk
		file, err := os.Open(filename)
		if err != nil {
			logs.Error("Failed to open file:", filename, err)
			// Optionally, skip the file and continue, or return an error response
			continue
		}
		defer file.Close()

		// Get the base filename (e.g., "document_A.pdf" from "static/downloads/document_A.pdf")
		// This ensures the files inside the zip don't have the full server path.
		baseFilename := filepath.Base(filename)

		// Create a file header within the zip archive
		header := &zip.FileHeader{
			Name:   baseFilename,
			Method: zip.Deflate, // Use Deflate for compression
		}

		// Create the writer for the file within the zip archive
		fileWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			logs.Error("Failed to create zip header for:", filename, err)
			continue
		}

		// Copy the content of the file into the zip file writer
		_, err = io.Copy(fileWriter, file)
		if err != nil {
			logs.Error("Failed to copy file content:", filename, err)
			continue
		}
	}

	// 4. (Implicitly handled by defer zipWriter.Close())
	// The zip stream is sent directly to the client as the files are added.
	// No need to load the entire archive into memory.

}

// FolderHtag Sélection d'un tag d'un album /folder/:beedirid/htagid
func (c *MainController) FolderTag() {
	user_id := c.GetSession("user_id").(string)

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	htagid := c.Ctx.Input.Param(":htagid")

	parent := beeDir.GetParent()

	// Sélection des sous-dossiers accessibles du bdir courant
	beeDirs := parent.GetParentBeedirs()

	// Construction de la liste des beefiles
	// album et sous-dossiers concernés par le htag
	beeFiles := []models.BeeFile{}
	for _, bdir := range *beeDirs {
		if bdir.ParentID == beeDir.ID || bdir.ID == beeDir.ID {
			for _, bfile := range bdir.BeeFiles {
				if slices.Contains(bfile.Keywords, htagid) {
					beeFiles = append(beeFiles, *bfile)
				}
			}
		}
	}
	// tri des beefiles
	sort.Slice(beeFiles, func(i, j int) bool {
		return beeFiles[i].DateOriginal < beeFiles[j].DateOriginal
	})

	c.Data["parent"] = &parent
	c.Data["beedirs"] = &beeDirs
	c.Data["beedir"] = &parent
	c.Data["beefiles"] = &beeFiles
	c.Data["htagid"] = htagid
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	c.SetSession("folder", c.Ctx.Request.RequestURI)

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "folder.html"
}

// Modifier un données metadata de l'image
func (c *MainController) Meta() {
	user_id := c.GetSession("user_id").(string)

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	beeFile := beeDir.BeeFiles[c.Ctx.Input.Param(":beefileid")]
	parent := beeDir.GetParent()

	flash := beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {

		// URL de url_return
		url_return := c.GetString("return")

		// ENREGISTREMENT DE L'IMAGE si modifiée
		simage := c.GetString("image")
		if len(simage) > 0 {
			err := beeFile.UpdateImage(simage)
			if err != nil {
				logs.Error(err)
				flash.Error("Beerama.UpdateImage %s", err)
				flash.Store(&c.Controller)
				c.Ctx.Redirect(302, url_return)
			}
		}

		// ENREGISTREMENT DU DOCUMENT
		texte := c.GetString("doc")
		if len(texte) > 0 {
			// maj de l'image avec canvas et du texte xml
			canvas := c.GetString("canvas")
			if len(canvas) > 0 {
				beeFile.CreateDocThumbnail(models.Config.Width, texte, canvas)
			}

		}

		// MAJ de beefile

		// title
		title := c.GetString("title")
		beeFile.Title = title
		// description
		description := c.GetString("description")
		beeFile.Description = description
		// Date Time Original
		dateoriginal := c.GetString("dateoriginal")
		beeFile.DateOriginal = dateoriginal
		timeoriginal := c.GetString("timeoriginal")
		beeFile.TimeOriginal = timeoriginal
		// Year
		year := c.GetString("year")
		beeFile.Year = year
		if year != "" {
			beeFile.DateOriginal = ""
			beeFile.TimeOriginal = ""
		}
		// image de couverture
		if c.GetString("couverture") == "yes" {
			beeFile.Source = "yes"
			// si déjà couverture valorisée -> maf du beefile concerné
			if parent.Couverture == "" {
				// pas de couverture actuellement
				parent.Couverture = beeFile.ID
			} else {
				// maj de l'ancienne couverture
				if parent.Couverture != beeFile.ID {
					bfile := models.GetBeeFile(parent.Couverture)
					bfile.Source = ""
					bfile.WriteMeta()
				}
				parent.Couverture = beeFile.ID
			}
		}
		// keywords
		keywords := c.GetStrings("keywords")
		beeFile.Keywords = keywords
		// raz de la date
		razdate := c.GetString("razdate")
		if razdate == "yes" {
			beeFile.Year = ""
			beeFile.DateOriginal = ""
			beeFile.TimeOriginal = ""
		}
		// urlexterne qui sera mémorisé dans exif.Subject
		urlExterne := c.GetString("urlexterne")
		if urlExterne != "" {
			if strings.Contains(urlExterne, "openstreetmap") {
				latitude, longitude := models.GetLatitudeLongitude(urlExterne)
				beeFile.UrlExterne = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=15/%s/%s&layers=P", latitude, longitude, latitude, longitude)
			} else {
				beeFile.UrlExterne = urlExterne
			}
		} else {
			beeFile.UrlExterne = ""
		}

		// report des meta dans le fichier
		err := beeFile.WriteMeta()
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, url_return)
		}

		beeDir.UpdateAlbum()
		// réindexation des beefiles
		models.Config.IndexAllBeefiles()
		c.Ctx.Redirect(302, url_return)
	}

	// Remplissage du contexte pour le template
	c.Data["parent"] = parent
	c.Data["beedir"] = &beeDir
	c.Data["beefile"] = &beeFile
	c.Data["htagid"] = ""
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	// cas des images drawio

	if beeFile.IsDrawio {
		// 1. Read the SVG file into a byte slice
		svgBytes, err := os.ReadFile(beeFile.Path)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama %s", err)
			flash.Store(&c.Controller)
		}
		// 2. Base64 encode the byte slice
		base64Encoded := base64.StdEncoding.EncodeToString(svgBytes)
		// 3. Prepend the Data URI scheme for SVG
		var dataURI string
		if beeFile.IsSvg {
			dataURI = "data:image/svg+xml;base64," + base64Encoded
		} else {
			dataURI = "data:image/png;base64," + base64Encoded
		}
		c.Data["content"] = template.URL(dataURI)
		// png vide
		// template.URL("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVQYV2NgYAAAAAMAAWgmWQ0AAAAASUVORK5CYII=")
	}
	if beeFile.IsDoc {
		html, err := models.GetMetaData(beeFile.Path, "CaptionWriter")
		if err != nil {
			c.Data["content"] = ""
		} else {
			c.Data["content"] = html
		}
	}

	// Mémorisation du dernier appel si folder
	if strings.Contains(c.Ctx.Request.RequestURI, "/folder/") {
		c.SetSession("folder", c.Ctx.Request.RequestURI)
	}

	c.TplName = "meta.html"
}

// Ajout d'un hahtag à l'album courant /e/tag/beedirid/beefileid
func (c *MainController) Tag() {
	user_id := c.GetSession("user_id").(string)

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	beeFile := beeDir.BeeFiles[c.Ctx.Input.Param(":beefileid")]

	beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {
		// AJOUT DU TAG
		keyword := strings.ToLower(c.GetString("keyword"))
		// maj
		parent := beeDir.GetParent()
		parent.KeywordsAlbum = append(parent.Keywords, keyword)
		beeDir.Keywords = append(beeDir.Keywords, keyword)
		beeDir.KeywordsAlbum = append(beeDir.KeywordsAlbum, keyword)
	}

	// actualisation
	c.Data["beedir"] = &beeDir
	c.Data["beefile"] = &beeFile
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	c.Ctx.Redirect(302, "/e/meta/"+beeDir.ID+"/"+beeFile.ID)
}

// Viewer document de type doc quill
func (c *MainController) Doc() {
	user_id := c.GetSession("user_id").(string)

	flash := beego.ReadFromRequest(&c.Controller)

	beeFile := models.GetBeeFile(c.Ctx.Input.Param(":beefileid"))
	beeDir := models.GetBeeDir(beeFile.DirID)

	html, err := models.GetMetaData(beeFile.Path, "CaptionWriter")
	if err != nil {
		flash.Error("%v", err)
		flash.Store(&c.Controller)
	}

	// Remplissage du contexte pour le template
	c.Data["content"] = &html
	c.Data["beedir"] = &beeDir
	c.Data["beefile"] = &beeFile
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	c.TplName = "doc.html"
}

// Retourne l'image
func (c *MainController) Document() {
	beeFile := models.GetBeeFile(c.Ctx.Input.Param(":beefileid"))
	// réponse directe
	// http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, beeFile.Path)

	w := c.Ctx.ResponseWriter
	if beeFile.IsDoc {
		if html, err := models.GetMetaData(beeFile.Path, "CaptionWriter"); err == nil {
			w.Header().Set("Content-Length", strconv.Itoa(len(html)))
			_, err = c.Ctx.ResponseWriter.Write([]byte(html))
			if err != nil {
				logs.Error("Erreur lors de l'écriture de la réponse:", err)
			}
		}
	} else {
		// 1. Lire le contenu du fichier
		imageData, err := os.ReadFile(beeFile.Path)
		if err != nil {
			http.Error(c.Ctx.ResponseWriter, "Impossible de lire l'image", http.StatusInternalServerError)
			return
		}

		// 2. Définir l'entête Content-Type
		// L'entête doit correspondre au format de l'image
		if beeFile.IsImage {
			if models.Contains([]string{".jpeg", ".jpg"}, strings.ToLower(beeFile.Ext)) {
				w.Header().Set("Content-Type", "image/jpg")
			} else if models.Contains([]string{".png"}, strings.ToLower(beeFile.Ext)) {
				w.Header().Set("Content-Type", "image/png")
			} else if models.Contains([]string{".gif"}, strings.ToLower(beeFile.Ext)) {
				w.Header().Set("Content-Type", "image/gif")
			} else if models.Contains([]string{".mp4"}, strings.ToLower(beeFile.Ext)) {
				w.Header().Set("Content-Type", "video/mp4")
			} else {
				w.Header().Set("Content-Type", "application/octet-stream")
				// http.Error(c.Ctx.ResponseWriter, "Not Image or vidéo", http.StatusInternalServerError)
			}
		}

		// Vous pouvez également définir Content-Length si vous connaissez la taille
		w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))

		// 3. Écrire les données de l'image dans la réponse
		_, err = c.Ctx.ResponseWriter.Write(imageData)
		if err != nil {
			logs.Error("Erreur lors de l'écriture de la réponse:", err)
		}
	}
}

// Restauration de l'image avec son original
func (c *MainController) Restore() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	beeFile := beeDir.BeeFiles[c.Ctx.Input.Param(":beefileid")]

	beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {
		beeFile.RestoreOriginal()
	}
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, "/e/meta/"+beeDir.ID+"/"+beeFile.ID)
}

// FileUpload Charger le fichier sur le serveur
func (c *MainController) Upload() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	flash := beego.ReadFromRequest(&c.Controller)

	files, err := c.GetFiles("files")
	if err != nil {
		goto Erreur
	}
	for _, mfile := range files {

		file, err := mfile.Open()
		if err != nil {
			goto Erreur
		}
		defer file.Close()

		logs.Info("Upload %v", mfile.Filename)

		fileContents, err := io.ReadAll(file)
		if err != nil {
			goto Erreur
		}
		path := beeDir.Path + "/" + mfile.Filename
		err = os.WriteFile(path, fileContents, 0644)
		if err != nil {
			goto Erreur
		}
		beeFile, err := beeDir.CreateBeeFile(path, true)
		if err == nil {
			flash.Notice("Le fichier %s a été ajouté", beeFile.Path)
			flash.Store(&c.Controller)
		} else {
			goto Erreur
		}

	}

	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
	return
Erreur:
	logs.Error(err)
	flash.Error("Beerama.Upload %s", err)
	flash.Store(&c.Controller)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// FileRm Supprimer le fichier ou dossier
func (c *MainController) FileRm() {
	// liste des fichiers à supprimer séparés avec des ,
	paths := strings.Split(c.GetString("paths"), ",")
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	flash := beego.ReadFromRequest(&c.Controller)

	// Suppression des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(path)

		err := beeFile.DeleteImage(beeDir)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Rm %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
	}
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// MkFolder Création d'un album
func (c *MainController) MkFolder() {
	newDir := c.GetString("new_name")
	path := models.Config.Racine + "/" + c.GetString("new_name")

	flash := beego.ReadFromRequest(&c.Controller)

	err := os.MkdirAll(path, 0744)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Mkdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, "/")
	}
	models.Config.AddFolder(newDir)
	c.Ctx.Redirect(302, "/")
}

// FolderRename
func (c *MainController) FolderRename() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	newName := c.GetString("new_name")

	flash := beego.ReadFromRequest(&c.Controller)

	// beeDir.UpdatePathBeeDir()
	err := beeDir.RenameBeeDir(newName)
	if err != nil {
		logs.Error(err)
		flash.Error("FolderRename %s", err)
		flash.Store(&c.Controller)
	}
	beeDir.UpdateAlbum()
	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// Document
func (c *MainController) NewDoc() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	title := c.GetString("new_name")

	flash := beego.ReadFromRequest(&c.Controller)

	pathSrc := "./static/modeles/quill.png"
	pathDest := beeDir.Path + "/" + models.GenerateKey() + ".doc.png"
	// copy du fichier source dans la destination
	err := shutil.CopyFile(pathSrc, pathDest, false)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Mkdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile, err := beeDir.CreateBeeFile(pathDest, true)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama CreateBeeFile %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile.Title = title
	beeFile.WriteMeta()

	// Rechargement de l'album
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// NewDraw
func (c *MainController) NewDraw() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	title := c.GetString("new_name")

	flash := beego.ReadFromRequest(&c.Controller)

	pathSrc := "./static/modeles/modele.drawio.png"
	pathDest := beeDir.Path + "/" + models.GenerateKey() + ".drawio.png"
	// copy du fichier source dans la destination
	err := shutil.CopyFile(pathSrc, pathDest, false)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Mkdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile, err := beeDir.CreateBeeFile(pathDest, true)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama CreateBeeFile %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile.Title = title
	beeFile.WriteMeta()

	// Rechargement de l'album
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// NewUrl
func (c *MainController) NewUrl() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	url := c.GetString("url")

	flash := beego.ReadFromRequest(&c.Controller)

	pathSrc := "./static/modeles/modele.url"
	pathDest := beeDir.Path + "/" + models.GenerateKey() + ".url"
	// copy du fichier source dans la destination
	err := shutil.CopyFile(pathSrc, pathDest, false)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Mkdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile, err := beeDir.CreateBeeFile(pathDest, true)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama CreateBeeFile %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFile.UrlImage = url
	beeFile.WriteMeta()

	// Rechargement de l'album
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// MkSubFolder Création d'un sous-dossier
func (c *MainController) MkSubFolder() {

	beedir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	newdir := c.GetString("new_name")
	path := models.Config.Racine + "/" + beedir.Name + "/" + newdir

	flash := beego.ReadFromRequest(&c.Controller)

	err := os.MkdirAll(path, 0744)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Mkdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, "/")
	}
	models.Config.AddSubFolder(beedir, newdir)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// Rmdir suppression d'un album ou sous-dossier
func (c *MainController) Rmdir() {

	beedir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	flash := beego.ReadFromRequest(&c.Controller)

	err := os.RemoveAll(beedir.Path)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Rmdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, "/")
	}
	models.Config.RemoveFolder(beedir)

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, "/")
}

// Rechargement de tout les albums et sous-dossiers
func (c *MainController) ReloadAll() {

	beego.ReadFromRequest(&c.Controller)

	models.LoadBeeDirs()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, "/")

}

// Rechargement de l'album
func (c *MainController) Reload() {
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	beego.ReadFromRequest(&c.Controller)

	beeDir.LoadBeeFiles()
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// Lot mise à jour des métadonnées des fichiers sélectionnés
func (c *MainController) Lot() {

	flash := beego.ReadFromRequest(&c.Controller)

	// liste des fichiers à dupliquerr séparés par des ,
	paths := strings.Split(c.GetString("paths"), ",")

	// récupération des champs à mettre à jour
	// title
	title := c.GetString("title")
	titleok := c.GetString("title_ok")
	// description
	description := c.GetString("description")
	descriptionok := c.GetString("description_ok")
	// Date Time Original
	dateoriginal := c.GetString("dateoriginal")
	dateoriginalok := c.GetString("dateoriginal_ok")
	timeoriginal := c.GetString("timeoriginal")
	timeoriginalok := c.GetString("timeoriginal_ok")
	// Year
	year := c.GetString("year")
	yearok := c.GetString("year_ok")
	if year != "" {
		dateoriginal = ""
		dateoriginalok = "yes"
		timeoriginal = ""
		timeoriginalok = "yes"
	}
	// keywords
	keywords := c.GetStrings("keywords")
	keywordsok := c.GetString("keywords_ok")
	// urlexterne
	urlexterne := c.GetString("urlexterne")
	urlexterneok := c.GetString("urlexterne_ok")
	if urlexterne != "" {
		if strings.Contains(urlexterne, "openstreetmap") {
			latitude, longitude := models.GetLatitudeLongitude(urlexterne)
			urlexterne = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=15/%s/%s&layers=P", latitude, longitude, latitude, longitude)
		}
	}

	// Traitement unitaire des fichiers
	// tri des fichiers
	sort.Slice(paths, func(i, j int) bool {
		return paths[1] < paths[j]
	})
	var dirid = ""
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(path)
		if titleok == "yes" {
			beeFile.Title = title
		}
		if descriptionok == "yes" {
			beeFile.Description = description
		}
		if dateoriginalok == "yes" {
			beeFile.DateOriginal = dateoriginal
		}
		if timeoriginalok == "yes" {
			beeFile.TimeOriginal = timeoriginal
		}
		if yearok == "yes" {
			beeFile.Year = year
		}
		if keywordsok == "yes" {
			beeFile.Keywords = keywords
		}
		if urlexterneok == "yes" {
			beeFile.UrlExterne = urlexterne
		}
		// report des meta dans le fichier
		err := beeFile.WriteMeta()
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		if dirid == "" {
			dirid = beeFile.DirID
		}
		if dirid != beeFile.DirID {
			models.GetBeeDir(beeFile.DirID).UpdateAlbum()
			dirid = beeFile.DirID
		}
	}
	models.GetBeeDir(dirid).UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// Duplicate Copier de(s) fichier(s) dans l'album répertoire courant
func (c *MainController) Duplicate() {
	// album source
	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	// liste des fichiers à dupliquer séparés par des ,
	paths := strings.Split(c.GetString("paths"), ",")

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(path)
		pathDest := beeDir.Path + "/cp_" + beeFile.Base
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Duplicate %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		_, err := beeDir.CreateBeeFile(pathDest, true)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Duplicate %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		// err = beeFileDuplicate.BackupImage()
		// if err != nil {
		// 	logs.Error(err)
		// 	flash.Error("Beerama.Upload %s", err)
		// 	flash.Store(&c.Controller)
		// 	c.Ctx.Redirect(302, c.GetSession("folder").(string))
		// }
	}
	beeDir.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// FileCopy Copier de(s) fichier(s) dans un autre album
func (c *MainController) FileCopy() {
	// album destination
	beeDirDest := models.Config.BeeDirs[c.GetString("destid")]

	// liste des fichiers à déplacer sépârés avec des ,
	paths := strings.Split(c.GetString("paths"), ",")

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(path)
		pathDest := beeDirDest.Path + "/cp_" + beeFile.Base
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			goto Erreur
		}
		beeFileDest, err := beeDirDest.CreateBeeFile(pathDest, true)
		if err != nil {
			goto Erreur
		}
		err = beeFileDest.BackupImage()
		if err != nil {
			goto Erreur
		}
	}
	beeDirDest.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
	return
Erreur:
	logs.Error(err)
	flash.Error("Beerama.Upload %s", err)
	flash.Store(&c.Controller)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// FileMove Déplacer le fichier
func (c *MainController) FileMove() {

	// liste des fichiers à déplacer séparés avec des ,
	paths := strings.Split(c.GetString("paths"), ",")
	// Répertoire destination
	beeDirDest := models.Config.BeeDirs[c.GetString("destid")]

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	// tri des fichiers
	sort.Slice(paths, func(i, j int) bool {
		return paths[1] < paths[j]
	})
	var dirid = ""
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(path)
		pathDest := beeDirDest.Path + "/" + beeFile.Base
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			goto Erreur
		}
		beeFileDest, err := beeDirDest.CreateBeeFile(pathDest, false)
		if err != nil {
			goto Erreur
		}
		err = beeFileDest.BackupImage()
		if err != nil {
			goto Erreur
		}
		err = beeFile.DeleteImage(models.GetBeeDir(beeFile.DirID))
		if err != nil {
			goto Erreur
		}
		if dirid == "" {
			dirid = beeFile.DirID
		}
		if dirid != beeFile.DirID {
			models.GetBeeDir(beeFile.DirID).UpdateAlbum()
			dirid = beeFile.DirID
		}
	}
	models.GetBeeDir(dirid).UpdateAlbum()
	beeDirDest.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
	return
Erreur:
	logs.Error(err)
	flash.Error("Beerama.FileMove %s", err)
	flash.Store(&c.Controller)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// DragDrop Glisser Déplacer un fichier dans un autre répertoire
func (c *MainController) DragDrop() {
	// paramètre action
	beeDirDest := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	// champs transmis
	dsrc := c.GetString("dsrc") // répertoire id
	fsrc := c.GetString("fsrc") // fichier id

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	beeDirSrc := models.Config.BeeDirs[dsrc]
	beefileSrc := beeDirSrc.BeeFiles[fsrc]
	pathDest := beeDirDest.Path + "/" + beefileSrc.Base
	if beefileSrc.Path == pathDest {
		err := errors.New("le déplacement d'une diapo dans le même répertoire est ignoré")
		logs.Error(err)
		flash.Error("drag drop %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	// copy du fichier source dans la destination
	err = shutil.CopyFile(beefileSrc.Path, pathDest, false)
	if err != nil {
		logs.Error(err)
		flash.Error("drag drop %s : %s -> %s", err, beefileSrc.Path, pathDest)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	beeFileDest, err := beeDirDest.CreateBeeFile(pathDest, false)
	if err != nil {
		logs.Error(err)
		flash.Error("drag drop %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	err = beeFileDest.BackupImage()
	if err != nil {
		logs.Error(err)
		flash.Error("drag drop %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	err = beefileSrc.DeleteImage(beeDirSrc)
	if err != nil {
		logs.Error(err)
		flash.Error("drag drop %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}

	beeDirDest.UpdateAlbum()
	beeDirSrc.UpdateAlbum()

	// réindexation des beefiles
	models.Config.IndexAllBeefiles()

	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// Search
func (c *MainController) Search() {
	user_id := c.GetSession("user_id").(string)

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]
	parent := beeDir.GetParent()
	// Sélection des sous-dossiers accessibles du bdir courant
	beeDirs := parent.GetParentBeedirs()

	var search string

	if c.Ctx.Input.Method() == "POST" {
		search = c.GetString("search")
		// Mémorisation du texte recherché dans la session
		c.SetSession("search", search)
	} else {
		search = c.Data["search"].(string)
	}

	flash := beego.ReadFromRequest(&c.Controller)

	if len(search) == 0 {
		c.DelSession("search")
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}

	s, err := fulltext.NewSearcher(models.Config.IndexDirs + "/idxout")
	if err != nil {
		logs.Error(err)
		flash.Error("Recherche %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}
	defer s.Close()

	// recherche des items dans l'index
	// et création d'un tableau des beefiles concernés
	sr, err := s.SimpleSearch(search, 20)
	if err != nil {
		logs.Error(err)
		flash.Error("Recherche %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, c.GetSession("folder").(string))
	}

	beeFiles := []models.BeeFile{}
	for _, item := range sr.Items {
		dirid, fileid, found := strings.Cut(string(item.Id), "_")
		if found {
			// on ne prend que les bdirs de l'album parent
			for _, bdir := range *beeDirs {
				if bdir.ID == dirid {
					beeFile := models.Config.BeeDirs[dirid].BeeFiles[fileid]
					beeFiles = append(beeFiles, *beeFile)
				}
			}
		}
	}

	// tri des beefiles
	sort.Slice(beeFiles, func(i, j int) bool {
		return beeFiles[i].DateOriginal < beeFiles[j].DateOriginal
	})

	c.Data["parent"] = &parent
	c.Data["beedirs"] = &beeDirs
	c.Data["beedir"] = &parent
	c.Data["beefiles"] = &beeFiles
	c.Data["search"] = search
	c.Data["htagid"] = ""
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	c.TplName = "folder.html"
}

// Modifier le fichier des users
func (c *MainController) Users() {

	flash := beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {

		content := c.GetString("content")

		// ENREGISTREMENT du fichier
		err := models.UpdateUsers([]byte(content))
		if err != nil {
			logs.Error(err)
			flash.Error("Users %s", err)
			flash.Store(&c.Controller)
		}
	}

	content, err := models.GetUsersContent()
	if err != nil {
		flash.Error("%v", err)
		flash.Store(&c.Controller)
	}

	// Remplissage du contexte pour le template
	c.Data["content"] = &content

	c.TplName = "users.html"
}

// Modifier le fichier des .beeaccess.conf d'un album
func (c *MainController) Access() {
	user_id := c.GetSession("user_id").(string)

	beeDir := models.Config.BeeDirs[c.Ctx.Input.Param(":beedirid")]

	flash := beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {

		content := c.GetString("content")

		// ENREGISTREMENT du fichier
		err := beeDir.UpdateAccess([]byte(content))
		if err != nil {
			logs.Error(err)
			flash.Error("Access %s", err)
			flash.Store(&c.Controller)
		}
	}

	content, err := beeDir.GetAccessContent()
	if err != nil {
		flash.Error("%v", err)
		flash.Store(&c.Controller)
	}

	// Remplissage du contexte pour le template
	c.Data["beedir"] = &beeDir
	c.Data["content"] = &content
	c.Data["is_editor"] = beeDir.IsUserEditor(user_id)

	c.TplName = "access.html"
}
