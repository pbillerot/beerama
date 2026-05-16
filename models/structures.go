package models

import (
	"os"
)

// Utilisateurs lus dans .beeusers.conf
type User struct {
	Password string `toml:"password"`
	IsAdmin  bool   `toml:"is_admin"`
	IsEditor bool   `toml:"is_editor"`
}
type Access struct {
	IsEditor bool `toml:"is_editor"`
}

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
	Racine      string              // chemin du répertoire racine
	Original    string              // chemin du répertoire des originaux
	Trash       string              // chemin de la corbeille
	Thumbnail   string              // chemin du répertoire racine des miniatures
	UsersPath   string              // chemin d'accès au fichier des utilisateurs
	Width       int                 // largeur de la vignette
	Height      int                 // hauteur de la vignette
	BeeDirs     map[string]*BeeDir  // Liste des BeeDir trouvés dans app.conf.beedir
	IndexDirs   string              // répertoire des index des mots fulltext
	Index       []byte              // index des mots full text
	Users       map[string]User     // Les utilisateurs
	BeeFiles    map[string]*BeeFile // index des beefiles
}

// Bee Context webapp courante dans la session
type BeeDir struct {
	Name          string              // nom du répertoire
	ID            string              // calculé par LoadBeeDirs
	Path          string              // chemin complet
	Title         string              // Titre de la répertoire trouve dans beemage.conf
	ParentID      string              // ID de l'album parent
	WithChildren  bool                // l'album possède de(s) sous-dossier(s)
	Count         int                 // nbre de diapos du dossier
	CountAlbum    int                 // nbre de diapos de l'album (tous les dossiers)
	BeeFiles      map[string]*BeeFile // la liste des fichiers de content
	Keywords      []string            // les hashtags du dossier
	KeywordsAlbum []string            // les hashtags de l'album
	Users         map[string]Access   // droit d'accès des utilisateurs
	Couverture    string              // id de l'image couverture de l'album
}

// BeeFile propriétés d'un fichier dans le sous-dossier BeeDir
type BeeFile struct {
	ID string // calculé par LoadBeeFiles
	// fichier
	Name       string // = Base sans l'extension
	Base       string
	Ext        string // extension du fichier
	Path       string // path de l'image calculé
	Categories string
	Date       string
	Dir        string
	// type
	IsAudio     bool
	IsDir       bool
	IsDrawio    bool
	IsDoc       bool
	IsImage     bool
	IsPdf       bool
	IsSvg       bool
	IsSystem    bool
	IsText      bool
	IsVideo     bool
	DirID       string // id du répertoire de l'image
	ParentID    string // id du répertoire parent du répertoire de l'image
	Original    string // path de l'original calculé
	Tags        string
	Thumb       string // chemin de la vignette
	UrlDocument string // url complete http
	UrlImage    string // uri de l'image
	UrlThumb    string // uri de la vignette
	UrlExterne  string // https://www.openstreetmap.org/?mlat=[Latitude]&mlon=[Longitude]#map=15/[Latitude]/[Longitude]
	// metadata
	Title        string // exif.Title
	Model        string
	Make         string
	Keywords     []string // exif.Kewords
	ISO          string
	ImageWidth   string
	ImageHeight  string
	FocalLength  string
	FileSize     string
	ExposureTime string
	Description  string // exif.Description
	Year         string // exif.Credit
	DateOriginal string // exif.DateOriginal
	TimeOriginal string // exif.TimeOriginal
	Altitude     string
	Latitude     string
	Longitude    string
	LensModel    string
	Source       string // =id pour inquer que l'image est la couverture de l'album
	Version      string // version de beerama dernière mise à jour
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

type FileUrl struct {
	Id               string `toml:"Id"`
	Title            string
	Description      string
	DateOriginal     string
	TimeOriginal     string
	Keywords         []string
	InternetShortcut struct {
		URL string
	}
}

// Metadata xml dans une image file.doc.png
type QuillXml struct {
	Html string `xml:"html"`
}
