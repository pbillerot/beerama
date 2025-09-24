package models

import (
	"os"
)

// BeeConfig structure du fichier de configuration de l'application app.conf
type BeeConfig struct {
	AppName     string
	Version     string
	Title       string // source dans beemage.conf de la racine
	Description string // source dans beemage.conf de la racine
	Background  string
	Favicon     string
	Icon        string
	Github      string
	Help        string
	HelpEditor  string
	Racine      string // chemin du répertoire racine
	Original    string // chemin du répertoire des originaux
	Thumbnail   string // chemin du répertoire racine des miniatures
	Width       uint   // largeur de la vignette
	Height      uint   // hauteur de la vignette
	// Liste des BeeDir trouvés dans app.conf.beedir
	BeeDirs []*BeeDir
}

// Bee Context webapp courante dans la session
type BeeDir struct {
	Name         string     // nom du répertoire
	ID           string     // calculé par LoadBeeDirs
	Path         string     // chemin complet
	Title        string     // Titre de la répertoire trouve dans beemage.yaml
	ParentID     string     // ID de l'album parent
	WithChildren bool       // l'album possède de(s) sous-dossier(s)
	Count        int        // nbre de photos du dossier
	CountAlbum   int        // nbre de photos de l'album (tous les dossiers)
	BeeFiles     []*BeeFile // la liste des fichiers de content
	Keywords     []string   // les hashtags de l'album
}

// BeeFile propriétés d'un fichier dans le sous-dossier BeeDir
type BeeFile struct {
	Action       string
	Base         string
	Categories   string
	Content      []byte
	Date         string
	Dir          string
	Ext          string // extension du fichier
	ID           string // calculé par LoadBeeFiles
	IsAudio      bool
	IsConf       bool
	IsDir        bool
	IsDrawio     bool
	IsExcel      bool
	IsImage      bool
	IsMarkdown   bool
	IsPdf        bool
	IsPowerpoint bool
	IsSystem     bool
	IsText       bool
	IsWord       bool
	Path         string // path de l'image calculé
	DirID        string // id du répertoire de l'image
	ParentID     string // id du répertoire parent du répertoire de l'image
	Original     string // path de l'original calculé
	Tags         string
	Title        string
	Thumb        string // chemin de la vignette
	UrlImage     string
	UrlThumb     string
	// metadata
	Model        string
	Make         string
	Keywords     []string
	ISO          string
	ImageWidth   string
	ImageHeight  string
	FocalLength  string
	FileSize     string
	ExposureTime string
	Description  string
	DateOriginal string
	TimeOriginal string
}

// BeePathInfo as
type BeePathInfo struct {
	Path string
	Info os.FileInfo
}

// Breadcrumb as
type Breadcrumb struct {
	Base   string
	Path   string
	IsLast bool
}
