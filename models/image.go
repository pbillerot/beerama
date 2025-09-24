package models

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/beego/beego/v2/core/logs"
	"github.com/nfnt/resize"
)

// RestoreOriginal
func (beeFile *BeeFile) RestoreOriginal() (err error) {

	// Lecture de l'image dans original
	file, err := os.Open(beeFile.Original)
	if err == nil {
		// 1. l'original existe
		sourceFile, err := os.Open(beeFile.Original)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		// 3. Create a destination file for writing.
		destinationFile, err := os.Create(beeFile.Path)
		if err != nil {
			return err
		}
		defer destinationFile.Close()
		// 4. Use io.Copy to copy data from the source to the destination.
		bytesCopied, err := io.Copy(destinationFile, sourceFile)
		if err != nil {
			return err
		}
		logs.Info("Image %s copied (%v bytes)", beeFile.Original, bytesCopied)
	} else {
		return err
	}
	defer file.Close()

	// mise à jour de la vignette
	err = beeFile.createThumbnail(Config.Width, Config.Height)

	// ENREGISTREMENT DES METADATA
	// Exiftool
	et, err := exiftool.NewExiftool()
	if err != nil {
		return
	}
	defer et.Close()
	originals := et.ExtractMetadata(beeFile.Path)
	originals[0].SetString("Description", beeFile.Description)
	// Date Time Original
	originals[0].SetString("DateTimeOriginal", beeFile.DateOriginal+" "+beeFile.TimeOriginal)
	// keywords
	keywords := beeFile.Keywords
	originals[0].SetStrings("Keywords", keywords)

	et.WriteMetadata(originals)

	return err
}

// AddBeeFile
// - création d'un beefile
// - recup des metadata
// - ajout du beefile dans beedir.beefiles
// - report de tout les hashtags des beefiles dans beedir
// - création du thumbnail
func (beeDir *BeeDir) AddBeeFile(path string, idstart int) (*BeeFile, error) {
	beeFile := &BeeFile{}
	// Exiftool
	// https://github.com/barasher/go-exiftool/tree/master
	et, err := exiftool.NewExiftool()
	if err != nil {
		return beeFile, err
	}
	defer et.Close()

	beeFile.Path = path
	beeFile.DirID = beeDir.ID
	// beeFile.ParentID = voir dans LoadBeeDirs
	beeFile.Dir = filepath.Dir(path)
	beeFile.Base = filepath.Base(path)
	beeFile.Ext = filepath.Ext(path)
	beeFile.UrlImage = "/album" + beeFile.Dir[len(Config.Racine):] + "/" + beeFile.Base

	if contains([]string{".jpeg", ".jpg", ".png"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
	} else if contains([]string{".conf"}, beeFile.Ext) {
		var content []byte
		content, err = os.ReadFile(beeFile.Path)
		if err != nil {
			logs.Error(err)
		}
		beeFile.Content = content
		beeFile.IsConf = true
	} else {
		beeFile.IsSystem = true
	}

	fileInfos := et.ExtractMetadata(beeFile.Path)
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}
		for k, v := range fileInfo.Fields {
			switch k {
			case "Model":
				beeFile.Model = v.(string)
			case "Make":
				beeFile.Make = v.(string)
			case "Keywords":
				beeFile.Keywords = beeFile.Keywords[:0]
				switch v := v.(type) {
				case string:
					kw := strings.Split(strings.ToLower(v), ",")
					beeFile.Keywords = append(beeFile.Keywords, kw...)
				case float64:
					beeFile.Keywords = append(beeFile.Keywords, fmt.Sprintf("%v", v))
				default:
					for _, vv := range v.([]any) {
						switch t := vv.(type) {
						case string:
							kw := strings.Split(strings.ToLower(vv.(string)), ",")
							beeFile.Keywords = append(beeFile.Keywords, kw...)
						case float64:
							beeFile.Keywords = append(beeFile.Keywords, strings.ToLower(fmt.Sprintf("%v", vv.(float64))))
						default:
							fmt.Printf("Type inconnu : %T pour %v", t, v)
						}
					}
				}
			case "ISO":
				beeFile.ISO = fmt.Sprintf("%v", v.(float64))
			case "ImageWidth":
				beeFile.ImageWidth = fmt.Sprintf("%v", v.(float64))
			case "ImageHeight":
				beeFile.ImageHeight = fmt.Sprintf("%v", v.(float64))
			case "FocalLength":
				beeFile.FocalLength = v.(string)
			case "FileSize":
				beeFile.FileSize = v.(string)
			case "ExposureTime":
				switch v := v.(type) {
				case string:
					beeFile.ExposureTime = v
				case float64:
					beeFile.ExposureTime = fmt.Sprintf("%v", v)
				default:
					beeFile.ExposureTime = fmt.Sprintf("%v", v)
				}
			case "Description":
				beeFile.Description = v.(string)
			case "DateTimeOriginal":
				beeFile.DateOriginal = strings.Replace(v.(string), ":", "-", 2)[:10]
				beeFile.TimeOriginal = v.(string)[11:16]
			}
		}
	}
	// Chemin Original et thumbnail
	dirOriginal := Config.Original + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	beeFile.Original = dirOriginal + beeFile.Base
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	beeFile.Thumb = dirThumb + "th_" + beeFile.Base
	beeFile.UrlThumb = "/thumb" + dirThumb[len(Config.Thumbnail):] + "th_" + beeFile.Base

	// ajout dans BeeFiles
	beeFile.ID = "id" + strconv.Itoa(idstart+len(beeDir.BeeFiles)+1)
	beeDir.BeeFiles = append(beeDir.BeeFiles, beeFile)

	// report des keywords dand beeDir
	beeDir.Keywords = append(beeDir.Keywords, beeFile.Keywords...)

	// création de la miniature dans Config.Thumbnail si n'existe pas
	if !beeFile.existeThumbnail() {
		beeFile.createThumbnail(Config.Width, Config.Height)
	}

	// logs.Info(beeDir.ID, beeFile.ID, beeFile.Path)

	return beeFile, nil
}

// updateMeta
func (beeFile *BeeFile) UpdateMeta() (err error) {

	// Exiftool
	// https://github.com/barasher/go-exiftool/tree/master
	et, err := exiftool.NewExiftool()
	if err != nil {
		return err
	}
	defer et.Close()
	originals := et.ExtractMetadata(beeFile.Path)
	originals[0].SetString("Description", beeFile.Description)
	// Date Time Original
	originals[0].SetString("DateTimeOriginal", beeFile.DateOriginal+" "+beeFile.TimeOriginal)
	// keywords
	keywords := beeFile.Keywords
	originals[0].SetStrings("Keywords", keywords)

	et.WriteMetadata(originals)

	return nil
}

// DeleteImage
// backup dans dossier des originals (une seule fois)
// suppression du fichier image
// mise à jour du beefiles du dossier source
func (beeFile *BeeFile) DeleteImage(beeDir *BeeDir) (err error) {
	// backup
	err = beeFile.BackupImage()
	if err != nil {
		return err
	}
	// delete
	err = os.RemoveAll(beeFile.Path)
	if err != nil {
		return
	}
	// suppression du beeFile de beeDir.BeeFiles
	// recherche de l'indice dans le tableau
	for index, file := range beeDir.BeeFiles {
		if file.Path == beeFile.Path {
			beeDir.BeeFiles = append(beeDir.BeeFiles[:index], beeDir.BeeFiles[index+1:]...)
			break
		}
	}
	return nil
}

// BackupImage
// backup dans dossier des originals (une seule fois)
func (beeFile *BeeFile) BackupImage() error {

	// calcul du répertoire destination
	dirPath := Config.Original + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	beeFile.Original = dirPath + beeFile.Base
	perm := os.FileMode(0755)

	// création des répertoires intermédiaires
	err := os.MkdirAll(dirPath, perm)
	if err != nil {
		return err
	}

	// Lecture de l'image dans original
	file, errexiste := os.Open(beeFile.Original)
	if errexiste != nil {
		// l'original n'existe pas -> backup
		sourceFile, err := os.Open(beeFile.Path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()
		// 3. Create a destination file for writing.
		destinationFile, err := os.Create(beeFile.Original)
		if err != nil {
			return err
		}
		defer destinationFile.Close()
		// 4. Use io.Copy to copy data from the source to the destination.
		bytesCopied, err := io.Copy(destinationFile, sourceFile)
		if err != nil {
			return err
		}
		logs.Info("Image %s copied (%v bytes)", beeFile.Original, bytesCopied)
	}
	defer file.Close()

	return err
}

// updateImage et backup dans dossier des orginals (une seule fois)
func (beeFile *BeeFile) UpdateImage(simage string) (err error) {

	err = beeFile.BackupImage()
	if err != nil {
		return err
	}

	b64data := simage[strings.IndexByte(simage, ',')+1:]
	unbased, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return err
	}
	err = os.WriteFile(beeFile.Path, unbased, 0644)
	if err != nil {
		return err
	}

	// mise à jour de la miniature
	err = beeFile.createThumbnail(Config.Width, Config.Height)
	return err
}

// existeThumbnail avec maj de beefile
func (beeFile *BeeFile) existeThumbnail() bool {
	_, err := os.Stat(beeFile.Thumb)
	return !os.IsNotExist(err)
}

// createThumbnail création de la vignette sous config.vignette
func (beeFile *BeeFile) createThumbnail(width, height uint) (err error) {
	// calcul et création des répertoires parents de la vignette
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	perm := os.FileMode(0755)
	err = os.MkdirAll(dirThumb, perm)
	if err != nil {
		return
	}

	// Lecture de l'image source
	file, err := os.Open(beeFile.Path)
	if err != nil {
		return
	}
	defer file.Close()

	var img image.Image

	if contains([]string{".png"}, strings.ToLower(beeFile.Ext)) {
		img, err = png.Decode(file)
	} else if contains([]string{".jpg", ".jpeg"}, strings.ToLower(beeFile.Ext)) {
		img, err = jpeg.Decode(file)
	} else {
		return
	}

	if err != nil {
		logs.Error("decode %s ", beeFile.Path)
		return
	}
	if img == nil {
		logs.Error("conversion %s ", beeFile.Path)
		return
	}
	// Resize the image to the specified width and height
	thumb := resize.Thumbnail(width, height, img, resize.Lanczos3)

	out, err := os.Create(beeFile.Thumb)
	if err != nil {
		logs.Error("create %s ", beeFile.Path)
		return
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, thumb, nil)

	logs.Info("Thumbnail créé %s ", beeFile.Thumb)
	return
}
