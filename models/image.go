package models

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/barasher/go-exiftool"
	"github.com/beego/beego/v2/core/logs"
	"github.com/disintegration/imaging"
	"github.com/pbillerot/beerama/shutil"
)

// Exiftool
const BUFFER_SIZE = 50 * 1024 * 1024 // 50MB

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
	buf := make([]byte, BUFFER_SIZE)
	et, err := exiftool.NewExiftool(exiftool.Buffer(buf, BUFFER_SIZE))
	if err != nil {
		return
	}
	defer et.Close()
	originals := et.ExtractMetadata(beeFile.Path)
	originals[0].SetString("Description", strings.ReplaceAll(beeFile.Description, "\n", "¤"))
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
func (beeDir *BeeDir) AddBeeFile(path string, newid bool) (*BeeFile, error) {
	beeFile := &BeeFile{}
	// Exiftool
	buf := make([]byte, BUFFER_SIZE)
	et, err := exiftool.NewExiftool(exiftool.Buffer(buf, BUFFER_SIZE))
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
	beeFile.UrlImage = "/s/album" + beeFile.Dir[len(Config.Racine):] + "/" + beeFile.Base

	if Contains([]string{".jpeg", ".jpg", ".png"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
		if strings.Contains(beeFile.Base, "drawio") {
			beeFile.IsDrawio = true
		}
	} else if Contains([]string{".gif"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
	} else if Contains([]string{".svg"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsImage = true
		beeFile.IsSvg = true
		if strings.Contains(beeFile.Base, "drawio") {
			beeFile.IsDrawio = true
		}
	} else if Contains([]string{".mov", ".m4v", ".mkv", ".mp4", ".webm"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsVideo = true
	} else if Contains([]string{".pdf"}, strings.ToLower(beeFile.Ext)) {
		beeFile.IsPdf = true
	} else if Contains([]string{".conf"}, beeFile.Ext) {
		var content []byte
		content, err = os.ReadFile(beeFile.Path)
		if err != nil {
			logs.Error(err)
		}
		beeFile.Content = content
		beeFile.IsConf = true
	} else if Contains([]string{".url"}, beeFile.Ext) {
		content, err := os.ReadFile(beeFile.Path)
		if err != nil {
			logs.Error(err)
		}
		fileurl := FileUrl{}
		err = toml.Unmarshal(content, &fileurl)
		if err != nil {
			logs.Error(err)
		}
		beeFile.Content = content
		beeFile.IsUrl = true
		beeFile.Description = fileurl.Description
		beeFile.DateOriginal = fileurl.DateOriginal
		beeFile.TimeOriginal = fileurl.TimeOriginal
		beeFile.Keywords = fileurl.Keywords
		beeFile.UrlImage = fileurl.InternetShortcut.URL
		beeFile.ID = fileurl.Id
	} else {
		beeFile.IsSystem = true
	}
	if beeFile.IsImage || beeFile.IsPdf || beeFile.IsVideo {
		fileInfos := et.ExtractMetadata(beeFile.Path)
		for _, fileInfo := range fileInfos {
			if fileInfo.Err != nil {
				fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
				continue
			}
			for k, v := range fileInfo.Fields {
				switch k {
				case "Title":
					switch t := v.(type) {
					case string:
						beeFile.ID = v.(string)
					default:
						beeFile.ID = ""
						logs.Error(beeFile.Base, t)
					}
				case "Model":
					beeFile.Model = v.(string)
				case "Make":
					beeFile.Make = v.(string)
				case "Keywords":
					beeFile.Keywords = beeFile.Keywords[:0]
					switch v := v.(type) {
					case string:
						kw := strings.Split(strings.ToLower(v), ",")
						j := 0 // Index to write non-empty strings
						for _, val := range kw {
							if len(val) > 0 {
								kw[j] = val
								j++
							}
						}
						// Reslice to the new length
						kw = kw[:j]
						beeFile.Keywords = append(beeFile.Keywords, kw...)
					case float64:
						beeFile.Keywords = append(beeFile.Keywords, fmt.Sprintf("%v", v))
					default:
						for _, vv := range v.([]any) {
							switch t := vv.(type) {
							case string:
								kw := strings.Split(strings.ToLower(vv.(string)), ",")
								j := 0 // Index to write non-empty strings
								for _, val := range kw {
									if len(val) > 0 {
										kw[j] = val
										j++
									}
								}
								// Reslice to the new length
								kw = kw[:j]
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
					beeFile.Description = strings.ReplaceAll(v.(string), "¤", "\n")
				case "DateTimeOriginal":
					if len(v.(string)) > 9 {
						beeFile.DateOriginal = strings.Replace(v.(string), ":", "-", 2)[:10]
					} else {
						beeFile.DateOriginal = ""
					}
					if len(v.(string)) > 15 {
						beeFile.TimeOriginal = v.(string)[11:16]
					} else {
						beeFile.TimeOriginal = ""
					}
				}
			}
		}
	}
	// Chemin Original et thumbnail
	dirOriginal := Config.Original + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	beeFile.Original = dirOriginal + beeFile.Base
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	if beeFile.IsPdf || beeFile.IsVideo {
		beeFile.Thumb = dirThumb + "th_" + beeFile.Base + ".jpg"
		beeFile.UrlThumb = "/s/thumb" + dirThumb[len(Config.Thumbnail):] + "th_" + beeFile.Base + ".jpg"

	} else {
		beeFile.Thumb = dirThumb + "th_" + beeFile.Base
		beeFile.UrlThumb = "/s/thumb" + dirThumb[len(Config.Thumbnail):] + "th_" + beeFile.Base
	}

	// ajout dans BeeFiles
	// reprise existant des document sans clé
	if newid {
		beeFile.ID = ""
	}
	beeFile.GetNewId()
	beeDir.BeeFiles[beeFile.ID] = beeFile
	// beeDir.BeeFiles = append(beeDir.BeeFiles, beeFile)

	// report des keywords dand beeDir
	beeDir.Keywords = append(beeDir.Keywords, beeFile.Keywords...)

	// création de la miniature dans Config.Thumbnail si n'existe pas
	if !beeFile.existeThumbnail() {
		beeFile.createThumbnail(Config.Width, Config.Height)
	}

	// Indexation du beefile
	beeFile.Idx()
	// logs.Info(beeDir.ID, beeFile.ID, beeFile.Path)

	return beeFile, nil
}

// updateMeta
func (beeFile *BeeFile) UpdateMeta() (err error) {
	if Contains([]string{".avi", ".mkv", ".m4v", ".ogv", ".webm"}, strings.ToLower(beeFile.Ext)) {
		// attention certains types ne sont pas modifiables
		// https://exiftool.org/exiftool_pod.html
		return fmt.Errorf("extension non modifiable: %s", beeFile.Base)
	}
	// Exiftool
	buf := make([]byte, BUFFER_SIZE)
	et, err := exiftool.NewExiftool(exiftool.Buffer(buf, BUFFER_SIZE))
	if err != nil {
		return err
	}
	defer et.Close()
	originals := et.ExtractMetadata(beeFile.Path)
	// id
	if originals[0].Err == nil {
		originals[0].SetString("Title", beeFile.GetNewId())
	} else {
		logs.Error(originals[0].Err)
	}
	// description
	if originals[0].Err == nil {
		originals[0].SetString("Description", strings.ReplaceAll(beeFile.Description, "\n", "¤"))
	} else {
		logs.Error(originals[0].Err)
	}
	// Date Time Original
	if originals[0].Err == nil {
		originals[0].SetString("DateTimeOriginal", beeFile.DateOriginal+" "+beeFile.TimeOriginal)
	} else {
		logs.Error(originals[0].Err)
	}
	// keywords
	if originals[0].Err == nil {
		if beeFile.IsPdf {
			originals[0].SetString("Keywords", strings.Join(beeFile.Keywords, ","))
		} else {
			// originals[0].SetStrings("Keywords", beeFile.Keywords)
			originals[0].SetString("Keywords", strings.Join(beeFile.Keywords, ","))
		}
	} else {
		logs.Error(originals[0].Err)
	}
	if Contains([]string{".avi", ".mkv", ".m4v", ".ogv", ".webm"}, strings.ToLower(beeFile.Ext)) {
		// attention certains types ne sont pas modifiables
		// https://exiftool.org/exiftool_pod.html
		logs.Warning("%s : metadata non modifiables", beeFile.Base)
	} else {
		if originals[0].Err == nil {
			et.WriteMetadata(originals)
		} else {
			logs.Error(originals[0].Err)
		}
	}

	return originals[0].Err
}

// DeleteImage
// backup dans dossier corbeille
// suppression du fichier image
func (beeFile *BeeFile) DeleteImage(beeDir *BeeDir) (err error) {
	// backup
	err = beeFile.TrashImage()
	if err != nil {
		return err
	}
	// suppression image
	err = os.RemoveAll(beeFile.Path)
	if err != nil {
		return
	}
	// suppression thumbnail
	err = os.RemoveAll(beeFile.Thumb)
	if err != nil {
		return
	}
	// suppression original
	err = os.RemoveAll(beeFile.Original)
	if err != nil {
		return
	}

	// suppression du beeFile de beeDir.BeeFiles
	delete(beeDir.BeeFiles, beeFile.ID)

	return nil
}

// Suppression d'une image (copie dans la corbeille)
func (beeFile *BeeFile) TrashImage() error {

	// calcul du répertoire destination
	dir := Config.Trash + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	dest := dir + beeFile.Base
	perm := os.FileMode(0755)

	// création des répertoires intermédiaires
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return err
	}

	// copie dans la corbeille
	err = shutil.CopyFile(beeFile.Path, dest, false)

	return err
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

	if beeFile.IsSvg {
		// decodeAndSaveSVG décode une Data URI Base64 en un fichier SVG.
		// 1. Définir le préfixe que nous attendons pour une image SVG Base64
		const prefix = "data:image/svg+xml;base64,"

		// 2. Vérifier si la chaîne commence par le préfixe attendu
		if !strings.HasPrefix(simage, prefix) {
			return fmt.Errorf("la chaîne Data URI n'est pas un format SVG Base64 valide : %s", simage[:20]+"...")
		}

		// 3. Extraire la partie Base64 pure (après le préfixe)
		base64Data := simage[len(prefix):]

		// 4. Décoder la chaîne Base64 en un tableau de bytes
		svgBytes, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return fmt.Errorf("échec du décodage Base64 : %w", err)
		}

		// 5. Écrire les bytes (qui sont le contenu XML du SVG) dans un fichier
		// Le mode 0644 donne au propriétaire les permissions de lecture/écriture, et lecture aux autres.
		err = os.WriteFile(beeFile.Path, svgBytes, 0644)
		if err != nil {
			return fmt.Errorf("échec de l'écriture du fichier SVG : %w", err)
		}
	} else {
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
	return err
}

// existeThumbnail avec maj de beefile
func (beeFile *BeeFile) existeThumbnail() bool {
	_, err := os.Stat(beeFile.Thumb)
	return !os.IsNotExist(err)
}

// createThumbnail création de la vignette sous config.vignette
func (beeFile *BeeFile) createThumbnail(width, _ int) (err error) {

	// 0. calcul et création des répertoires parents de la vignette
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	perm := os.FileMode(0755)
	err = os.MkdirAll(dirThumb, perm)
	if err != nil {
		return fmt.Errorf("error opening image: %s %w", beeFile.Path, err)
	}
	if beeFile.IsVideo {
		cmd := exec.Command("ffmpeg", "-i", beeFile.Path, "-ss", "00:00:01", "-vframes", "1", beeFile.Thumb)
		// Optional: Capture command output or errors
		// output, err := cmd.CombinedOutput()
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("failed to save thumbnail: %s %v", beeFile.Thumb, err)
			return err
		}
		StampOnImage(beeFile.Thumb, "./static/img/video-256.png")
	} else {
		// 1. Open the original image
		img, err := imaging.Open(beeFile.Path, imaging.AutoOrientation(true))
		if err != nil {
			err = fmt.Errorf("error opening image: %s %w", beeFile.Path, err)
			return err
		}
		// 2. Create the thumbnail
		// imaging.Thumbnail resizes the image to fit the specified dimensions
		// and crops the image to the exact size without distorting the aspect ratio.
		thumbnail := imaging.Resize(img, width, 0, imaging.CatmullRom)

		// 3. Save the thumbnail image to a file
		err = imaging.Save(thumbnail, beeFile.Thumb)
		if err != nil {
			err = fmt.Errorf("failed to save image: %s %v", beeFile.Path, err)
			return err
		}
	}

	logs.Info("Thumbnail créé %s ", beeFile.Thumb)
	return
}

func StampOnImage(pathImage, pathIcon string) (err error) {
	// 1. Charger l'image de base
	baseImageFile, err := os.Open(pathImage)
	if err != nil {
		log.Fatalf("Failed to open base image: %v", err)
	}
	defer baseImageFile.Close()

	baseImage, _, err := image.Decode(baseImageFile)
	if err != nil {
		log.Fatalf("Failed to decode base image: %v", err)
	}

	// 2. Charger l'icône
	iconFile, err := os.Open(pathIcon)
	if err != nil {
		log.Fatalf("Failed to open icon image: %v", err)
	}
	defer iconFile.Close()

	iconImage, _, err := image.Decode(iconFile)
	if err != nil {
		log.Fatalf("Failed to decode icon image: %v", err)
	}

	// 3. Créer une nouvelle image
	// La nouvelle image aura la même taille que l'image de base.
	// Nous utilisons un image.RGBA pour supporter la transparence si nécessaire.
	bounds := baseImage.Bounds()
	newImage := image.NewRGBA(bounds)

	// 4. Dessiner l'image de base sur la nouvelle image
	// Copie l'image de base sur la nouvelle image, en utilisant un opérateur de source pour copier simplement les pixels.
	draw.Draw(newImage, bounds, baseImage, image.Point{}, draw.Src)

	// 5. Dessiner l'icône sur la nouvelle image
	// Définissez la position de l'icône. Ici, nous la plaçons dans le coin supérieur gauche.
	// Vous pouvez ajuster iconX et iconY pour placer l'icône où vous le souhaitez.
	iconX := 10 // Décalage X depuis le coin supérieur gauche de l'image de base
	iconY := 10 // Décalage Y depuis le coin supérieur gauche de l'image de base

	// Calculer le rectangle de destination pour l'icône sur la nouvelle image
	iconBounds := iconImage.Bounds()
	iconDestRect := image.Rect(iconX, iconY, iconX+iconBounds.Dx(), iconY+iconBounds.Dy())

	// Dessiner l'icône. draw.Over est généralement utilisé pour superposer avec transparence.
	// Si l'icône n'a pas de transparence, draw.Src fonctionnera aussi.
	draw.Draw(newImage, iconDestRect, iconImage, image.Point{}, draw.Over)

	// 6. Sauvegarder la nouvelle image
	outputFile, err := os.Create(pathImage)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, newImage); err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	log.Println("Image with icon overlay saved as output_image.png")

	return err

}
