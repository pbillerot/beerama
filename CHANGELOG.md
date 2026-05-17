# CHANGELOG
- coloriage syntaxique :
- https://korben.info/quill-lediteur-wysiwyg-nouvelle-generation.html
- https://quilljs.com/docs/modules/syntax#syntax-highlighter-module

### À venir :
- vidage automatique de la poubelle des fichiers > 30 jours
x intégrer excalidraw (pas forcément utile car drawio)

2026.5.14
- `fixed` traitement des cas avec UserComment encore en html

2026.5.13
- `changed` suppression script de migration car traité en production

2026.5.12
- `added` "beerama <version>" enregistré dans metadata.CreatorTool
- `added` "migration 2026.5.12" des doc : metadata.CaptionWriter en json dans metadata.UserComment 
- `removed` suppression des metada.CaptionWriter si IsDoc

2026.5.11
- `added` Quill enregistre désormais dans UserComment en json (format natif de Quill) pour conserver le formatage des code-block
- `added` Coloriage syntaxique limitée à certains langages : 'plaintext', 'bash', 'css', 'javascript', 'json', 'python', 'sql', 'yaml', 'xml'
- `removed` suppression des isConf


2026.5.10
- `added` coloriage syntaxique dans le code des documents

2026.5.9
- `changed` upload zone orange pour déposer les fichiers

2026.5.8
- `fixed` bouton maj actif si tag ajouté

2026.5.7
- `fixed` tag en double suite maj meta
- `fixed` pas de title en fond orange sur le titre de la diapo si reandonly
- `changed` droppable en orange

2026.5.6
- `fixed` tag en double suite maj meta

2026.5.4
- `fixed` tag en double suite ajout

2026.5.3
- `changed` upload champ sur fond vert pour "dropper" les fichiers

2026.5.2
- `fixed` bug sélection d'une étiquette qui n'existe plus - clic sur couverture /none
- `fixed` url avec new pour les drawio
- `changed` icône document en bas à droite

2026.5.1
- `fixed` correction étiquettes vides ,,
- `changed` icône document et vidéo désormais à gauche car la référence était placé um même endroit

4.3.2 du 17 janv 2026
- `fixed` correction impossible enlever couverture

4.3.1 du 17 janv 2026
- `changed` menu avec la couverture
- `added` menu rétractable via une icône
- `changed` suite recherche texte ou par htag le nombre de diapos trouvées est affiché

4.2.0 du 13 janv 2026
- `added` script de purge de la corbeille
- `added` image et légende de couverture de l'album enregistré dans exif.Source

4.1.0 du 12 janv 2026
- `changed` retour à l'identification Basic

4.0.2 du 12 janv 2026
- `fixed` après une recherche et suite à l'édition d'une diapo, on reviendra toujours sur la page de recherche

4.0.2 du 11 janv 2026
- `changed` 5 1er caractères du nom de fichier en sur-impression des diapos
- `added` recheche sur ces 5 premiers caractères

4.0.1 du 4 janv 2026
- `fixed` filtre de lecture de l'authentification pour tous les chemains

4.0.0 du 4 janv 2026
- `changed` identification authentification confiées à authélia

3.7.2 du 3 janv 2026
- `changed` index en minuscules et sans accent

3.7.1 du 2 janv 2026
- `added` bouton raz sur meta
- `added` bouton undo sur meta

3.7.0 du 2 janv 2026
- `added` bouton raz sur meta
- `added` bouton undo sur meta

3.6.2 du 1er janv 2026
- `fixed` date et urlexterne non maj par lot

3.6.1 du 1er janv 2026
- `fixed` onchange sur checkbox non pris en compte
- `fixed` si annee valorisée raz date et heure

3.6.0 du 1er janv 2026
- `added` mises à jour de métadonnées par lot
- `fixed` correction liens map et document dans meta.html

3.5.0 du 1er janv 2026
- `added` 2 boutons séparés pour copier coller les meta dans le localStorage
- `added` raz localStorage automatique à l'entrée de Beerama

3.4.0 du 31 déc 2025
- `added` copier coller dans et du presse-papier
- `added` année seule ou complete

3.3.4 du 29 déc 2025
- `fixed` gestion erreur maj metadata
- `fixed` ajout nouvelle étiquette faisait perdre les autres

3.3.3 du 29 déc 2025
- `fixed` bee-press en double qui créé un fichier undefined

3.3.2 du 29 déc 2025
- `fixed` bee-press correction perte du nom de fichier

3.3.1 du 28 déc 2025
- `changed` menu gauche toujousr fixe même sur smartphone
- `added` metadata exif.LensModel

3.3.0 du 25 déc 2025
- `fixed` clic sur image non pris en compte sur tablette - conflit avec darg & drop
- `fixed` titre initialisé avec le nom du fichier au départ (si vide)
- `changed` le type url n'est plus pris en compte
- `added` drag & drop seulement sur le titre de l'image

3.2.2 du 5 déc 2025
- `fixed` opération de copie qui recréait le même identifiant
- `fixed` sous-dossiers non visualisés suite rename d'un sous-dossier

3.2.1 du 3 déc 2025
- `fixed` lien index non actualisé suite drag & drop

3.2.0 du 3 déc 2025
- `fixed` retour home après connexion refusée lien actif sur forbidden
- `fixed` erreurs dans logs
- `added` nouveau type de document "doc" via l'éditeur javascript Quill
- `added` doc Quill mémorisé dans exif.CaptionWriter

3.1.18 du 30 nov 2025
- `fixed` Subject à blanc si Urlexterne non renseignée

3.1.18 du 30 nov 2025
- `added` notice log avec n°ip

3.1.17 du 29 nov 2025
- `fixed` bug image url externe toujours valorisée

3.1.16 du 27 nov 2025
- `fixed` icone acces à l'url du document pour le partager éventuellemnt (l'url est copieée aussi dans le presse-papier)
- `fixed` suite édition d'une image le bouton "enregistrer et fermer" n'était pas actif

3.1.15 du 26 nov 2025
- `fixed` liste des sous-dossiers absents dans album
- `fixed` url clipboard

3.1.14 du 26 nov 2025
- `added` bouton lien vers copier le lien complet du document
- `fixed` liste nulle des albums cibles d'un déplacement

3.1.13 du 25 nov 2025
- `changed` rename de urlOSM en urlExterne pour fusionner avec les fichiers url

3.1.12 du 17 nov 2025
- `fixed` bug renommage d'un album

3.1.11 du 14 nov 2025
- `fixed` mode info en production

3.1.10 du 14 nov 2025
- `fixed` correction renommage systématique des fichiers lors du rechargement d'un dossier

3.1.9 du 12 nov 2025
- `removed` raz des données GPS non prise en compte par exif (ou alors c'est plus compliqué)

3.1.8 du 12 nov 2025
- `changed` url openstreetmap dans exif.Subject au lieu de exif.Comment (car Comment non traité dans pdf)

3.1.7 du 12 nov 2025
- `fixed` url openstreetmap mieux contrôlée

3.1.6 du 11 nov 2025
- `added` saisie directe de l'url openstreetmap pour localiser le lieu de la prise de vue

3.1.5 du 11 nov 2025
- `fixed` suppression étiquettes en doublon après un uplaod

3.1.4 du 11 nov 2025
- `fixed` relecture des metadata après restauration
- `fixed` correction chemin des vignettes des vidéos
- `changed` meta: raz des données gps

3.1.3 du 10 nov 2025
- `fixed` correction update meta systématique

3.1.2 du 10 nov 2025
- `fixed` drawio: prise en compte du nom.drawio.png

3.1.1 du 10 nov 2025
- `changed` url: affichage du titre dans la vignette

3.1.0 du 9 nov 2025
- `added` prise en compt des données GPS et affichage de la prise de vue dans OmpenStreetMap

3.0.3 du 8 nov 2025
- `fixed` suppression des étiquettes en double

3.0.2 du 8 nov 2025
- `fixed` lors des uploads, renommage du fichier systématique pour éviter doublons

3.0.1 du 7 nov 2025
- `changed` labels sous l'image du diaporama

3.0.0 du 7 nov 2025
- `added` fichiers renommés en aa-00-00-00
- `added` titre de la photo saisissable
- `changed` recherche sur titre, description, étiquettes, model de l'appareil photo, nom du ficher sans l'extension 

2.2.3 du 4 nov 2025
- `added` .gif traité
- `changed` les fichiers non traités ou extension inconnue sont affichés pour suppression éventuelle

2.2.2 du 4 nov 2025
- `changed` ouverture url dans nouvelle fenêtre

2.2.1 du 4 nov 2025
- `added` supprimer un album / admin
- `changed` lightbox pour visualiser la vidéa dans l'écram meta

2.2.0 du 3 nov 2025
- `added` gestion des vidéos mp4 webm mov m4v mkv
- `changed` toolbar drawio plus simple

2.1.2 du 2 nov 2025
- `fixed` nouvelle étiquette ne fonctiannait plus
- `fixed` nouvel album opérationnel de nouveau
- `fixed` étiquettes qui ne s'affichaient pas
- `changed` bouton à gauche dans meta
- `changed` pas de légende pour pdf drawio url

2.1.1 du 1er nov. 2025
- `fixed` création d'un album beedir.beefiles non initialisée

2.1.0 du 1er nov. 2025
- `changed` id file en clé unique
- `changed` viewer pdf plus large

2.0.4 du 29 oct. 2025
- `fixed` nouvelle étiquette ne fonctiannait plus
- `fixed` nouvel album opérationnel de nouveau
- `fixed` étiquettes qui ne s'affichaient pas
- `changed` bouton à gauche dans meta
- `changed` pas de légende pour pdf drawio url

2.0.3 du 29 oct. 2025
- `fixed` compteur des albums à zéro
- `changed` optimisation calcul compteurs et vignettes

2.0.2 du 28 oct. 2025
- `fixed` keywords à blanc

2.0.1 du 28 oct. 2025
- `fixed` correction création sous-dossier (beefiles non initialisées)

2.0.0 du 27 oct. 2025
- `added` gestion des users dans beeusers.conf
- `added` gestion des droits dans beeaccess.conf
- `added` fichier.url pour gérer des web application externes

1.8.1 du 25 oct. 2025
- `fixed` drag & drop incohérent sur dossiers
- `added` htag new lors d'import de fichiers ou de nouveau dessin
- `added` visualiseur pdf intégré avec appel propriétés

1.8.0 du 25 oct. 2025
- `added` gestion des accès aux albums lecture/écriture via le fichier beeaccess.yaml
- `added` ajout action download de diapo
- `added` ajout action de renommage d'une diapo

1.7.1 du 20 oct. 2025
- `added` protection des albums et vignettes du site 

1.7.0 du 19 oct. 2025
- `changed` url changée pour gérer les acces editor /e/ admin /a/
- `added` éditeur du fichier users.yaml
- `changed` retour à la ligne acceptée dans champ description

1.6.1 du 17 oct. 2025
- `fixed` correction message orange qui ne s'effacait pas sur abandon upload
- `added` menu et actions en fonction du profil de l'utilisateur
- `added` menu fonction de l'écran: si < 768px mobile sinon desktop ou tablette

1.6.0 du 17 oct. 2025
- `added` ajout d'un contrôle basique de connection user / password défini dans app.conf.users

1.5.7 du 15 oct. 2025
- `changed` changement de module pour créer les vignettes qui n'étaient pas toujours bien orientées

1.5.6 du 11 oct. 2025
- `changed` description lightbox sur fond noir

1.5.5 du 11 oct. 2025
- `changed` remplacement de lightbox par glightbox

1.5.4 du 10 oct. 2025
- `fixed` correction retour sur validation par touche return

1.5.3 du 10 oct. 2025
- `changed` enregistrement des keywords en string "h1, h2"

1.5.2 du 10 oct. 2025
- `changed` la librairie unipdf non libre a été changée par pdf.js (merci mozilla)

1.5.1 du 9 oct. 2025
- `fixed` changement de librairie pdf

1.5.0 du 9 oct. 2025
- `added` intégration des pdfs dans les albums
- `changed` généralisataion de la taille du buffer de exiftool à 50M 

1.4.0 du 9 oct. 2025
- `added` intégration de l'éditeur de dessin drawio
- `added` menu création d'un nouveau dessin
- `added` placeholder dans modal new
- `fixed` retout home si pas de contexte folder
- `fixed` augmentation du buffer des metadata

1.3.1 du 7 oct. 2025
- `added` bouton Enregistrer et Retour dans l'écran meta
- `changed` mémorisation adresse de retour lors de l'appel de meta, pourrevir sur une page mots-clés par exemple

1.3.0 du 6 oct. 2025
- `added` moteur de recherche dans commentaire htag et appareil https://github.com/bradleypeabody/fulltext

1.2.0 du 3 oct. 2025
- `added` répertoire corbeille pour recevoir les images suprimmées

1.1.5 du 29 sept. 2025
- `changed` page d'accueil image centrée

1.1.4 du 27 sept. 2025
- `changed` menu fixe is_admin

1.1.3 du 26 sept. 2025
- `addged` mode lecteur.écriture
- `changed` menu standard fixe

1.1.2 du 24 sept. 2025
- `added` sur diapo si sélection htag lien vers le dossier
- `added` formulaire metadata avec coche pour raz de la date et heure
1.1.1 du 24 sept. 2025
- `added` ajout compteur de photos dans le dossier
- `changed` le renommage d'un dossier n'entraîne plus le rechargement de tous les albums
- `changed` étiquettes en minuscule

1.1.0 du 23 sept. 2025
- `added` menu à gauche de l'album et de ses sous dossiers

1.0.6 du 22 sept. 2025
- `fixed` favicon en png
- `fixed` crossorigin du manifest

1.0.5 du 22 sept. 2025
- `fixed` manifest.pwa erreur chemin
- `fixed` metadata ExposureTime en float64
- `added` simplificationchargement des dossiers des albums

1.0.4 du 22 sept. 2025
- `fixed` bug affichage diapo sur tablette (attente que toutes les images soient chargées)
- `added` manifest.pwa Une PWA se consulte comme un site web classique, depuis une URL sécurisée mais permet une expérience utilisateur similaire à celle d'une application mobile

1.0.3 du 22 sept. 2025
- `fixed` bug affichage diapo sur tablette (une seule colonne)

1.0.2 du 22 sept. 2025
- `fixed` affichage diapo corrigé sur tablette 
- `added` lightbox retour au début si à la fin, gestion touch device (tablette)
- `added` drag and drop sur tablette
- `fixed` dockerfile avec exiftool
- `fixed` dockerfile avec exiftool

1.0.1 du 21 sept. 2025
- `fixed` dockerfile avec exiftool

1.0.0 du 21 sept. 2025
- `fixed` des keywords des sous-dossiers partagés dans l'album
- `fixed` prise en compte des keywords avec séparateurs virgule
- `added` paramétrage pour docker pour le site de production
- `added` mémorisation de la dernière diapo en édition

0.8.0 du 20 sept. 2025
- `fixed` glisser déposer ok sur sélection par htag
- `added` sélection par htag dans l'album et sous-dossiers
- `added` htag commun à l'album et sous dossier
- `fixed` renommage des albums et sous-dossiers ok en rechargeant completement la structure

0.7.0 du 19 sept. 2025
- fonction de renommage des répertoire non fonctionnelle
- `fixed` message flash corrigé
- `added` création des nouvelles thumbnails seulement au démarrage
- `fixed` retour meta mémorisation de l'url folder
- `fixed` lors sélection htag d'un sous-dossier perte barre des sous-dossiers
- `fixed` nouveau tag n'était plus enregistré dans la beedir

0.6.0 du 18 sept. 2025
- `changed` script jquery et autre dans static
- `added` glisser déplacer dans les sous-dossiers
- `added` suppression d'un album ou dossier si vide
- `added` création d'un album ou dossier
- `added` duplication d'une diapo dans le même album

0.5.0 du 15 sept. 2025
- `added` fonction de copier déplacer dans un autre album
- `added` prise en compte des png
- `added` projet beemage renommer en beerama
- `added` fonction de rechargement d'un album avec message wait (nag)
- `added` upload d'images dans l'album
- `added` restauration de l'original
- `added` suppression des images sélectionnées

0.4.0 du 10 sept. 2025
- `added` utilisation de filerobot pour modifier les images
- `added` sauvegarde des originaux dans un répertoire défini dans app.conf
- `changed` tri des images sur la date original
- `added` utilisation de lightbox pour visualiser les images en diaporama

0.3.0 du 9 septembre 2025
- `added` ihm avec les hashtags

0.2.0 du 7 septembre 2025
- recup metadata DateTimeOriginal Title Description Keywords en entre autres

0.1.0 du 1er septembre 2025
- `changed` fomantic 2.9.4 jquery 3.7.1 masonry

0.0.1 du 29 août 2025
- `changed` rename victor en beemage
- `removed` nettoyage go.mod go.sum .git public

###### Types de changements:
`added` *pour les nouvelles fonctionnalités.*  
`changed` *pour les changements aux fonctionnalités préexistantes.*  
`deprecated` *pour les fonctionnalités qui seront bientôt supprimées*.  
`removed` *pour les fonctionnalités désormais supprimées.*  
`fixed` *pour les corrections de bugs.*  
`security` *en cas de vulnérabilités.*  
