package controllers

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/beego/beego/v2/core/logs"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/pbillerot/beerama/models"
	"github.com/pbillerot/beerama/shutil"
)

// Main as get and Post
func (c *MainController) Main() {

	beego.ReadFromRequest(&c.Controller)

	c.Data["beedir"] = models.BeeDir{}

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "index.html"
}

// Folder Sélection d'un folder à administrer /folder/:beedirid
func (c *MainController) Folder() {

	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	c.Data["parent"] = models.GetBeeDir(beeDir.ParentID)
	c.Data["beedir"] = &beeDir
	c.Data["htagid"] = ""

	// Mémorisation du dernier appel
	c.SetSession("folder", c.Ctx.Request.RequestURI)

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "index.html"
}

// FolderHtag Sélection d'un folder à administrer /folder/:beedirid/htagid
func (c *MainController) FolderHtag() {

	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	htagid := c.Ctx.Input.Param(":htagid")
	fusion := *beeDir // clonage
	// fusion des beefiles de l'album et des sous-dossiers
	for _, bdir := range models.Config.BeeDirs {
		if bdir.ParentID == beeDir.ID {
			fusion.BeeFiles = append(fusion.BeeFiles, bdir.BeeFiles...)
		}
	}
	// tri des beefiles
	sort.Slice(fusion.BeeFiles, func(i, j int) bool {
		return fusion.BeeFiles[i].DateOriginal < fusion.BeeFiles[j].DateOriginal
	})

	c.Data["parent"] = models.GetBeeDir(beeDir.ParentID)
	c.Data["beedir"] = &fusion
	c.Data["htagid"] = htagid

	c.SetSession("folder", c.Ctx.Request.RequestURI)

	beego.ReadFromRequest(&c.Controller)

	c.TplName = "index.html"
}

// Modifier un données metadata de l'image
func (c *MainController) Meta() {
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	beeFile := models.GetBeeFile(c.Ctx.Input.Param(":beedirid"), c.Ctx.Input.Param(":beefileid"))

	flash := beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {

		// ENREGISTREMENT DE L'IMAGE si modifiée
		simage := c.GetString("image")
		if len(simage) > 0 {
			err := beeFile.UpdateImage(simage)
			if err != nil {
				logs.Error(err)
				flash.Error("Beerama.Upload %s", err)
				flash.Store(&c.Controller)
				c.Ctx.Redirect(302, "/meta/"+beeDir.ID+"/"+beeFile.ID)
			}
		}

		// MAJ de beefile

		// description
		description := c.GetString("description")
		beeFile.Description = description
		// Date Time Original
		dateoriginal := c.GetString("dateoriginal")
		beeFile.DateOriginal = dateoriginal
		timeoriginal := c.GetString("timeoriginal")
		beeFile.TimeOriginal = timeoriginal
		// keywords
		keywords := c.GetStrings("keywords")
		beeFile.Keywords = keywords
		// raz de la date
		razdate := c.GetString("razdate")
		if razdate == "on" {
			beeFile.DateOriginal = ""
			beeFile.TimeOriginal = ""
		}

		// report des meta dans l'image
		err := beeFile.UpdateMeta()
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		beeDir.UpdateBeeDir()
	}

	// Remplissage du contexte pour le template
	c.Data["parent"] = models.GetBeeDir(beeDir.ParentID)
	c.Data["beedir"] = &beeDir
	c.Data["beefile"] = &beeFile

	c.TplName = "meta.html"
}

// Ajout d'un hahtag à l'album courant /tag/beedirid/beefileid
func (c *MainController) Tag() {
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	beeFile := models.GetBeeFile(c.Ctx.Input.Param(":beedirid"), c.Ctx.Input.Param(":beefileid"))

	beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {
		// AJOUT DU TAG
		keyword := strings.ToLower(c.GetString("keyword"))
		// maj du beefile
		beeFile.Keywords = append(beeFile.Keywords, keyword)
		// report des keywords dans BeeDir sans doublons et triés
		beeDir.AddKeyword(keyword)
	}

	// actualisation
	c.Data["beedir"] = &beeDir
	c.Data["beefile"] = &beeFile

	c.Ctx.Redirect(302, "/meta/"+beeDir.ID+"/"+beeFile.ID)
}

// Restauration de l'image avec son original
func (c *MainController) Restore() {
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	beeFile := models.GetBeeFile(c.Ctx.Input.Param(":beedirid"), c.Ctx.Input.Param(":beefileid"))

	beego.ReadFromRequest(&c.Controller)

	if c.Ctx.Input.Method() == "POST" {
		beeFile.RestoreOriginal()
	}
	beeDir.UpdateBeeDir()
	c.Ctx.Redirect(302, "/meta/"+beeDir.ID+"/"+beeFile.ID)
}

// FileUpload Charger le fichier sur le serveur
func (c *MainController) Upload() {
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))

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
		fileContents, err := io.ReadAll(file)
		if err != nil {
			goto Erreur
		}
		path := beeDir.Path + "/" + mfile.Filename
		err = os.WriteFile(path, fileContents, 0644)
		if err != nil {
			goto Erreur
		}
		beeFile, err := beeDir.AddBeeFile(path, 0)
		if err == nil {
			flash.Notice("L'image %s a été ajoutée", beeFile.Path)
			flash.Store(&c.Controller)
		} else {
			goto Erreur
		}

	}
	beeDir.UpdateBeeDir()

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
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))

	flash := beego.ReadFromRequest(&c.Controller)

	// Suppression des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(beeDir, path)

		err := beeFile.DeleteImage(beeDir)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Rm %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
	}
	beeDir.UpdateBeeDir()
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
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	newName := c.GetString("new_name")

	flash := beego.ReadFromRequest(&c.Controller)

	// beeDir.UpdatePathBeeDir()
	err := beeDir.RenameBeeDir(newName)
	if err != nil {
		logs.Error(err)
		flash.Error("FolderRename %s", err)
		flash.Store(&c.Controller)
	}
	// Rechargement de albums
	beeDir.LoadBeeFiles(0)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// MkSubFolder Création d'un sous-dossier
func (c *MainController) MkSubFolder() {

	beedir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
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

	beedir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	flash := beego.ReadFromRequest(&c.Controller)

	err := os.RemoveAll(beedir.Path)
	if err != nil {
		logs.Error(err)
		flash.Error("Beerama Rmdir %s", err)
		flash.Store(&c.Controller)
		c.Ctx.Redirect(302, "/")
	}
	models.Config.RemoveFolder(beedir)
	c.Ctx.Redirect(302, "/")
}

// Rechargement de tout les albums et sous-dossiers
func (c *MainController) ReloadAll() {

	beego.ReadFromRequest(&c.Controller)

	models.LoadBeeDirs()
	c.Ctx.Redirect(302, "/")

}

// Rechargement de l'album
func (c *MainController) Reload() {
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))

	beego.ReadFromRequest(&c.Controller)

	beeDir.LoadBeeFiles(0)
	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// Duplicate Copier de(s) fichier(s) dans un autre album
func (c *MainController) Duplicate() {
	// album source
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))

	// liste des fichiers à dupliquerr séparés par des ,
	paths := strings.Split(c.GetString("paths"), ",")

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(beeDir, path)
		pathDest := beeDir.Path + "/cp_" + beeFile.Base
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Upload %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		beeFileDuplicate, err := beeDir.AddBeeFile(pathDest, 0)
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Upload %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
		err = beeFileDuplicate.BackupImage()
		if err != nil {
			logs.Error(err)
			flash.Error("Beerama.Upload %s", err)
			flash.Store(&c.Controller)
			c.Ctx.Redirect(302, c.GetSession("folder").(string))
		}
	}
	beeDir.UpdateBeeDir()
	c.Ctx.Redirect(302, c.GetSession("folder").(string))

}

// FileCopy Copier de(s) fichier(s) dans un autre album
func (c *MainController) FileCopy() {
	// album source
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	// album destination
	beeDirDest := models.GetBeeDir(c.GetString("destid"))

	// liste des fichiers à déplacer sépârés avec des ,
	paths := strings.Split(c.GetString("paths"), ",")

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(beeDir, path)
		var pathDest string
		if beeDir.ID == beeDirDest.ID {
			pathDest = beeDirDest.Path + "/cp_" + beeFile.Base
		} else {
			pathDest = beeDirDest.Path + "/" + beeFile.Base
		}
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			goto Erreur
		}
		beeFileDest, err := beeDirDest.AddBeeFile(pathDest, 0)
		if err != nil {
			goto Erreur
		}
		err = beeFileDest.BackupImage()
		if err != nil {
			goto Erreur
		}
	}
	beeDirDest.UpdateBeeDir()
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
	// uri
	beeDir := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))

	// liste des fichiers à déplacer séparés avec des ,
	paths := strings.Split(c.GetString("paths"), ",")
	// Répertoire destination
	beeDirDest := models.GetBeeDir(c.GetString("destid"))

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	// Traitement unitaire des fichiers
	for _, path := range paths {
		beeFile := models.GetBeeFilePath(beeDir, path)
		pathDest := beeDirDest.Path + "/" + beeFile.Base
		// copy du fichier source dans la destination
		err = shutil.CopyFile(beeFile.Path, pathDest, false)
		if err != nil {
			goto Erreur
		}
		beeFileDest, err := beeDirDest.AddBeeFile(pathDest, 0)
		if err != nil {
			goto Erreur
		}
		err = beeFileDest.BackupImage()
		if err != nil {
			goto Erreur
		}
		err = beeFile.DeleteImage(beeDir)
		if err != nil {
			goto Erreur
		}
	}
	beeDir.UpdateBeeDir()
	beeDirDest.UpdateBeeDir()
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
	beeDirDest := models.GetBeeDir(c.Ctx.Input.Param(":beedirid"))
	// champs transmis
	dsrc := c.GetString("dsrc") // répertoire id
	fsrc := c.GetString("fsrc") // fichier id

	flash := beego.ReadFromRequest(&c.Controller)
	var err error
	beefileSrc := models.GetBeeFile(dsrc, fsrc)
	beeDirSrc := models.GetBeeDir(dsrc)
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
	beeFileDest, err := beeDirDest.AddBeeFile(pathDest, 0)
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

	beeDirDest.UpdateBeeDir()
	beeDirSrc.UpdateBeeDir()
	c.Ctx.Redirect(302, c.GetSession("folder").(string))
}

// Mode Administration des albume
func (c *MainController) Admin() {

	beego.ReadFromRequest(&c.Controller)

	if boolValue, ok := c.GetSession("is_admin").(bool); ok {
		// boolValue is now the Go boolean true or false
		if boolValue {
			// is admin ok
			c.SetSession("is_admin", false)
		} else {
			// is admin ko
			c.SetSession("is_admin", true)

		}
	} else {
		c.SetSession("is_admin", true)
	}

	c.Ctx.Redirect(302, "/")

}
