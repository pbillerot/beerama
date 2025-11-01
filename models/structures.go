package models

import (
	"os"
)

type User struct {
	Password string `toml:"password"`
	IsAdmin  bool   `toml:"is_admin"`
	IsEditor bool   `toml:"is_editor"`
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
	Racine      string             // chemin du répertoire racine
	Original    string             // chemin du répertoire des originaux
	Trash       string             // chemin de la corbeille
	Thumbnail   string             // chemin du répertoire racine des miniatures
	UsersPath   string             // chemin d'accès au fichier des utilisateurs
	Width       int                // largeur de la vignette
	Height      int                // hauteur de la vignette
	BeeDirs     map[string]*BeeDir // Liste des BeeDir trouvés dans app.conf.beedir
	IndexDirs   string             // répertoire des index des mots fulltext
	Index       []byte             // index des mots full text
	Users       map[string]User    // Les utilisateurs
}

// Bee Context webapp courante dans la session
type BeeDir struct {
	Name          string              // nom du répertoire
	ID            string              // calculé par LoadBeeDirs
	Path          string              // chemin complet
	Title         string              // Titre de la répertoire trouve dans beemage.conf
	ParentID      string              // ID de l'album parent
	WithChildren  bool                // l'album possède de(s) sous-dossier(s)
	Count         int                 // nbre de photos du dossier
	CountAlbum    int                 // nbre de photos de l'album (tous les dossiers)
	BeeFiles      map[string]*BeeFile // la liste des fichiers de content
	Keywords      []string            // les hashtags du dossier
	KeywordsAlbum []string            // les hashtags de l'album
	Users         map[string]Access
}
type Access struct {
	IsEditor bool `toml:"is_editor"`
}

// BeeFile propriétés d'un fichier dans le sous-dossier BeeDir
type BeeFile struct {
	ID           string // calculé par LoadBeeFiles
	Action       string
	Base         string
	Categories   string
	Content      []byte
	Date         string
	Dir          string
	Ext          string // extension du fichier
	IsAudio      bool
	IsConf       bool
	IsDir        bool
	IsDrawio     bool
	IsExcel      bool
	IsImage      bool
	IsMarkdown   bool
	IsPdf        bool
	IsPowerpoint bool
	IsSvg        bool
	IsSystem     bool
	IsText       bool
	IsUrl        bool
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

type FileUrl struct {
	Id               string `toml:"Id"`
	Description      string `toml:"Description"`
	DateOriginal     string
	TimeOriginal     string
	Keywords         []string
	InternetShortcut struct {
		URL string
	}
}
