package models

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"beerama/shutil"

	"github.com/barasher/go-exiftool"
	"github.com/beego/beego/v2/core/logs" // External package for chunk manipulation
	"github.com/disintegration/imaging"
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

	beeFile.GetMetadata()

	return err
}

// Fonction utilitaire pour afficher la chaîne nettoyée dans l'exemple
func GetCleanedGPS(s string) string {
	if strings.HasSuffix(s, "W") {
		// longitude négative
		s = "-" + s
	}
	s = strings.ReplaceAll(s, ",", ".")
	reg := regexp.MustCompile(`[\p{L}a-zA-Z ]`)
	return reg.ReplaceAllString(s, "")
}

func GetMetaData(path string, param string) (data string, err error) {
	// 1. Initialiser ExifTool avec l'option de format de coordonnées décimales
	buf := make([]byte, BUFFER_SIZE)
	// Format décimal avec 6 décimales
	et, err := exiftool.NewExiftool(exiftool.CoordFormant("%.6f"), exiftool.Buffer(buf, BUFFER_SIZE))
	if err != nil {
		logs.Error("Erreur lors de l'initialisation de go-exiftool: %v", err)
	}
	defer et.Close()

	// 2. Extraire les métadonnées
	fileMetadata := et.ExtractMetadata(path)

	if len(fileMetadata) == 0 {
		logs.Trace("Aucune métadonnée extraite pour le fichier : %s", path)
		return "", nil
	}

	metadata := fileMetadata[0]
	if value, err := metadata.GetString(param); err == nil {
		return value, err
	} else {
		logs.Warning("exif: %s %s %w", path, param, err)
	}

	return "", err
}

func SetMetaData(path string, param string, value string) (err error) {
	// 1. Initialiser ExifTool avec l'option de format de coordonnées décimales
	buf := make([]byte, BUFFER_SIZE)
	et, err := exiftool.NewExiftool(exiftool.Buffer(buf, BUFFER_SIZE))
	if err != nil {
		logs.Error("Erreur lors de l'initialisation de go-exiftool: %v", err)
	}
	defer et.Close()

	// 2. Extraire les métadonnées
	fileMetadata := et.ExtractMetadata(path)

	if len(fileMetadata) == 0 {
		logs.Trace("Aucune métadonnée extraite pour le fichier : %s", path)
		return nil
	}

	metadata := fileMetadata[0]
	if metadata.Err == nil {
		metadata.SetString(param, value)
	} else {
		logs.Error("error exif: %s %w", path, metadata.Err)
	}
	if metadata.Err == nil {
		et.WriteMetadata(fileMetadata)
	} else {
		logs.Error("exif set : %s %s %w", path, param, metadata.Err)
	}

	if metadata.Err != nil {
		logs.Error("error set string: %s %w", path, metadata.Err)
		return metadata.Err
	}
	return nil
}

// valorisation de beefile avec les metadata de l'image
func (beeFile *BeeFile) GetMetadata() (err error) {
	// 1. Initialiser ExifTool avec l'option de format de coordonnées décimales
	buf := make([]byte, BUFFER_SIZE)
	// Format décimal avec 6 décimales
	et, err := exiftool.NewExiftool(exiftool.CoordFormant("%.6f"), exiftool.Buffer(buf, BUFFER_SIZE))
	if err != nil {
		logs.Error("Erreur lors de l'initialisation de go-exiftool: %v", err)
	}
	defer et.Close()

	// 2. Extraire les métadonnées
	fileMetadata := et.ExtractMetadata(beeFile.Path)

	if len(fileMetadata) == 0 {
		logs.Trace("Aucune métadonnée extraite pour le fichier : %s", beeFile.Path)
	}

	metadata := fileMetadata[0]

	// GPS
	if value, err := metadata.GetString("GPSLatitude"); err == nil {
		beeFile.Latitude = GetCleanedGPS(value)
	}
	if value, err := metadata.GetString("GPSLongitude"); err == nil {
		beeFile.Longitude = GetCleanedGPS(value)
		beeFile.UrlExterne = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=15/%s/%s&layers=P", beeFile.Latitude, beeFile.Longitude, beeFile.Latitude, beeFile.Longitude)
	}
	if value, err := metadata.GetString("GPSAltitude"); err == nil {
		beeFile.Altitude = GetCleanedGPS(value)
	}
	if value, err := metadata.GetString("Subject"); err == nil {
		if strings.Contains(value, "openstreetmap") {
			latitude, longitude := GetLatitudeLongitude(value)
			beeFile.UrlExterne = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=15/%s/%s&layers=P", latitude, longitude, latitude, longitude)
		} else {
			beeFile.UrlExterne = value
		}
	}

	// titre et description
	if value, err := metadata.GetString("Title"); err == nil {
		beeFile.Title = value
	}
	if value, err := metadata.GetString("Description"); err == nil {
		beeFile.Description = value
	}
	// image
	if value, err := metadata.GetString("LensModel"); err == nil {
		beeFile.LensModel = value
	}
	if value, err := metadata.GetString("ISO"); err == nil {
		beeFile.ISO = value
	}
	if value, err := metadata.GetString("ImageWidth"); err == nil {
		beeFile.ImageWidth = value
	}
	if value, err := metadata.GetString("ImageHeight"); err == nil {
		beeFile.ImageHeight = value
	}
	if value, err := metadata.GetString("FocalLength"); err == nil {
		beeFile.FocalLength = value
	}
	if value, err := metadata.GetString("ExposureTime"); err == nil {
		beeFile.ExposureTime = value
	}
	if value, err := metadata.GetString("Credit"); err == nil {
		beeFile.Year = value
	}
	if value, err := metadata.GetString("Source"); err == nil {
		beeFile.Source = value
	}
	if value, err := metadata.GetString("DateTimeOriginal"); err == nil {
		if len(value) > 9 {
			beeFile.DateOriginal = strings.Replace(value, ":", "-", 2)[:10]
		} else {
			beeFile.DateOriginal = ""
		}
		if len(value) > 15 {
			beeFile.TimeOriginal = value[11:16]
		} else {
			beeFile.TimeOriginal = ""
		}
	}
	// fichier
	if value, err := metadata.GetString("FileSize"); err == nil {
		beeFile.FileSize = value
	}
	// étiquettes
	if value, err := metadata.GetString("Keywords"); err == nil {
		beeFile.Keywords = beeFile.Keywords[:0]
		beeFile.Keywords = append(beeFile.Keywords, strings.Split(value, ",")...)
	}

	return err
}

// updateMeta
func (beeFile *BeeFile) WriteMeta() (err error) {
	// logs.Trace("UpdateMeta", beeFile.Path)
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
	fileMetadata := et.ExtractMetadata(beeFile.Path)
	metadata := fileMetadata[0]
	// title
	if metadata.Err == nil {
		metadata.SetString("Title", beeFile.Title)
	} else {
		logs.Error(metadata.Err)
	}
	// description
	if metadata.Err == nil {
		metadata.SetString("Description", strings.ReplaceAll(beeFile.Description, "\n", "¤"))
	} else {
		logs.Error(metadata.Err)
	}
	// GPS
	if metadata.Err == nil {
		if beeFile.UrlExterne != "" {
			metadata.SetString("Subject", beeFile.UrlExterne)
		}
	} else {
		logs.Error(metadata.Err)
	}
	// Year
	if metadata.Err == nil {
		metadata.SetString("Credit", beeFile.Year)
	} else {
		logs.Error(metadata.Err)
	}
	// Source
	if metadata.Err == nil {
		metadata.SetString("Source", beeFile.Source)
	} else {
		logs.Error(metadata.Err)
	}
	// Date Time Original
	if metadata.Err == nil {
		metadata.SetString("DateTimeOriginal", beeFile.DateOriginal+" "+beeFile.TimeOriginal)
	} else {
		logs.Error(metadata.Err)
	}
	// keywords
	if metadata.Err == nil {
		keywords := cleanJoin(beeFile.Keywords, ",")
		if beeFile.IsPdf {
			metadata.SetString("Keywords", keywords)
		} else {
			// originals[0].SetStrings("Keywords", beeFile.Keywords)
			metadata.SetString("Keywords", keywords)
		}
	} else {
		logs.Error(metadata.Err)
	}
	if metadata.Err == nil {
		et.WriteMetadata(fileMetadata)
	} else {
		logs.Error(metadata.Err)
	}

	return fileMetadata[0].Err
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

// updateImage et backup dans dossier des originals (une seule fois)
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

func (beeFile *BeeFile) createThumbnail(width, _ int) (err error) {
	if beeFile.IsPdf {
		return nil
	}
	// 0. calcul et création des répertoires parents de la vignette
	dirThumb := Config.Thumbnail + beeFile.Path[len(Config.Racine):len(beeFile.Path)-len(beeFile.Base)]
	perm := os.FileMode(0755)
	err = os.MkdirAll(dirThumb, perm)
	if err != nil {
		logs.Error("error opening image: %s %w", dirThumb, err)
		return err
	}
	if beeFile.IsVideo {
		cmd := exec.Command("ffmpeg", "-i", beeFile.Path, "-ss", "00:00:01", "-vframes", "1", beeFile.Thumb)
		// Optional: Capture command output or errors
		// output, err := cmd.CombinedOutput()
		err = cmd.Run()
		if err != nil {
			logs.Error("error ffmpeg: %s %w", beeFile.Path, err)
			return err
		}
		StampOnImage(beeFile.Thumb, "./static/img/video-256.png")
	} else {
		// 1. Open the original image
		img, err := imaging.Open(beeFile.Path, imaging.AutoOrientation(true))
		if err != nil {
			logs.Error("error image open: %s %w", beeFile.Path, err)
			return err
		}
		// 2. Create the thumbnail
		// imaging.Thumbnail resizes the image to fit the specified dimensions
		// and crops the image to the exact size without distorting the aspect ratio.
		thumbnail := imaging.Resize(img, width, 0, imaging.Lanczos)

		// 3. Save the thumbnail image to a file
		err = imaging.Save(thumbnail, beeFile.Thumb)
		if err != nil {
			err = fmt.Errorf("failed to save image: %s %v", beeFile.Path, err)
			return err
		}
		if beeFile.IsDoc {
			StampOnImageRightBottom(beeFile.Thumb, "./static/img/doc.png")
		}
	}

	logs.Info("Thumbnail créé %s ", beeFile.Thumb)
	return
}

func decodeAndSavePNG(base64Str string, filename string) error {
	// 1. Clean the string if it contains a data URI header (e.g., "data:image/png;base64,")
	parts := strings.Split(base64Str, ",")
	encodedData := base64Str
	if len(parts) > 1 {
		encodedData = parts[1]
	}

	// 2. Decode the base64 string into a byte slice
	pngBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		logs.Error("error base64 encoding: %s %w", filename, err)
		return err
	}

	// 3. Write the raw byte slice to a file
	err = os.WriteFile(filename, pngBytes, 0644)
	if err != nil {
		logs.Error("error write: %s %w", filename, err)
		return err
	}

	// 4 réduction de la taille
	// par relecture puis cropping
	img, err := imaging.Open(filename)
	if err != nil {
		logs.Error("error image open: %s %w", filename, err)
		return err
	}
	// 2. Create the thumbnail
	// imaging.Thumbnail resizes the image to fit the specified dimensions
	// and crops the image to the exact size without distorting the aspect ratio.
	thumbnail := imaging.Crop(img, image.Rect(0, 0, Config.Width, Config.Height))
	// thumbnail := imaging.Resize(img, width, 0, imaging.Lanczos)

	// 3. Save the thumbnail image to a file
	err = imaging.Save(thumbnail, filename)
	if err != nil {
		err = fmt.Errorf("failed to save image: %s %v", filename, err)
		return err
	}
	StampOnImageRightBottom(filename, "./static/img/doc.png")

	// fmt.Printf("Successfully decoded and saved PNG to: %s\n", filename)
	return nil
}

// createDocThumbnail création d'une nouvelle image, vignette et html dans exif.CaptionWriter
func (beeFile *BeeFile) CreateDocThumbnail(width int, html, capture string) (err error) {
	// remplacement de l'image par la capture
	if err := decodeAndSavePNG(capture, beeFile.Path); err != nil {
		return err
	}
	// enregistrement des données json dans les metadata
	err = SetMetaData(beeFile.Path, "CaptionWriter", html)
	return err
}

// StampOnImageLeft en position 10 10
func StampOnImage(pathImage, pathIcon string) (err error) {
	// 1. Charger l'image de base
	baseImageFile, err := os.Open(pathImage)
	if err != nil {
		logs.Error("Failed to open base image: %v", err)
		return err
	}
	defer baseImageFile.Close()

	baseImage, _, err := image.Decode(baseImageFile)
	if err != nil {
		logs.Error("Failed to decode base image: %v", err)
		return err
	}

	// 2. Charger l'icône
	iconFile, err := os.Open(pathIcon)
	if err != nil {
		logs.Error("Failed to open icon image: %v", err)
		return err
	}
	defer iconFile.Close()

	iconImage, _, err := image.Decode(iconFile)
	if err != nil {
		logs.Error("Failed to decode icon image: %v", err)
		return err
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
	iconDestRect := image.Rect(iconX,
		iconY,
		iconX+iconBounds.Dx(),
		iconY+iconBounds.Dy(),
	)

	// Dessiner l'icône. draw.Over est généralement utilisé pour superposer avec transparence.
	// Si l'icône n'a pas de transparence, draw.Src fonctionnera aussi.
	draw.Draw(newImage, iconDestRect, iconImage, image.Point{}, draw.Over)

	// 6. Sauvegarder la nouvelle image
	outputFile, err := os.Create(pathImage)
	if err != nil {
		logs.Error("Failed to create output file: %v", err)
		return err
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, newImage); err != nil {
		logs.Error("Failed to encode image: %v", err)
	}

	return err

}

// StampOnImageLeft en position 10 à droite et en bas
func StampOnImageRightBottom(pathImage, pathIcon string) (err error) {
	// 1. Charger l'image de base
	baseImageFile, err := os.Open(pathImage)
	if err != nil {
		logs.Error("Failed to open base image: %v", err)
		return err
	}
	defer baseImageFile.Close()

	baseImage, _, err := image.Decode(baseImageFile)
	if err != nil {
		logs.Error("Failed to decode base image: %v", err)
		return err
	}

	// 2. Charger l'icône
	iconFile, err := os.Open(pathIcon)
	if err != nil {
		logs.Error("Failed to open icon image: %v", err)
		return err
	}
	defer iconFile.Close()

	iconImage, _, err := image.Decode(iconFile)
	if err != nil {
		logs.Error("Failed to decode icon image: %v", err)
		return err
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
	// en bas à droite
	iconBounds := iconImage.Bounds()
	iconDestRect := image.Rect(
		bounds.Dx()-iconBounds.Dx()-iconX,
		bounds.Dy()-iconY,
		bounds.Dx()-iconX,
		bounds.Dy()-iconBounds.Dy()-iconY,
	)
	// en haut à droire
	// iconDestRect := image.Rect(
	// 	bounds.Dx()-iconBounds.Dx()-iconX,
	// 	iconY,
	// 	bounds.Dx()-iconX,
	// 	iconY+iconBounds.Dy(),
	// )

	// Dessiner l'icône. draw.Over est généralement utilisé pour superposer avec transparence.
	// Si l'icône n'a pas de transparence, draw.Src fonctionnera aussi.
	draw.Draw(newImage, iconDestRect, iconImage, image.Point{}, draw.Over)

	// 6. Sauvegarder la nouvelle image
	outputFile, err := os.Create(pathImage)
	if err != nil {
		logs.Error("Failed to create output file: %v", err)
		return err
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, newImage); err != nil {
		logs.Error("Failed to encode image: %v", err)
	}

	return err

}

func cleanJoin(elements []string, sep string) string {
	var result []string
	for _, str := range elements {
		// On vérifie que la chaîne n'est pas vide après avoir retiré les espaces
		if trimmed := strings.TrimSpace(str); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, sep)
}
