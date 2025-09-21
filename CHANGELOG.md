
# CHANGELOG

### À venir :
- intégration draw.io
- jquery droppable draggable non fonctionnel sur tablette
- passez en mode administrateur via mot de passe dans la session

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
