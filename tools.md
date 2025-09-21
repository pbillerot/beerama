# BEEDULE DEVELOPPEMENT

## Installation de GO sur mon portable T480 sous Ubuntu 25.04
```sh
sudo apt install golang-go
```
### dans ~/.profile
```bash
# Personnalisation
export PATH=$PATH:/usr/local/go/bin:/home/billerot/go/bin
export GOPATH=/home/billerot/go
```
```sh
source $HOME/.profile
go version
```
go version go1.24.2 linux/amd64

## Ajout de modules

### BeeGo
- cd $GOPATH
- go install github.com/beego/beego/v2@latest
- go install github.com/beego/bee/v2@latest
ou pour maj
- go get -u github.com/beego/beego/v2
- go get -u github.com/beego/bee/v2

### Environnement de développement
- go mod init github.com/pbillerot/beemage
- go mod tidy
- go build
- ./beemage
- ou lancer le Debug dans vscodium
- ou bee run

### Mise en production
- maj changelog.md app.conf
- git push...
- sur vps cd /volshare/docker/beemage ./build.sh

### markdown
https://github.com/gomarkdown/markdown
https://github.com/russross/blackfriday
- go get -u github.com/gomarkdown/markdown

### Scheduler
- https://github.com/MichaelS11/go-scheduler

Pour traduire l'expression
- https://github.com/bradymholt/cRonstrue

### Mémo de commandes (pour le débutant que je suis)
- go mod init github.com/pbillerot/beemage
- go mod tidy pour mettre à jour go.mod
- go get ./...
- go get -u

 - git tag v1.0.0
 - git push --tags origin

### Google Analytics
https://www.soberkoder.com/google-analytics-Bee/

### Build a REST API in Golang
https://www.soberkoder.com/go-rest-api-gorilla-mux/

### Erreur compilation sqlite3
```
././c/sqlite3.c: In function 'sqlite3SelectNew':
././c/sqlite3.c:128048:10: warning: function may return address of local variable [-Wreturn-local-addr]
128048 |   return pNew;
       |          ^~~~
././c/sqlite3.c:128008:10: note: declared here
128008 |   Select standin;
       |          ^~~~~~~
```
- export CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

### GIT
``` Retour à un checkout particulier ou effacement de l'historique
https://www.grafikart.fr/formations/git/checkout-revert-reset
Permet de revenir en arrière jusqu'au <commit>, réinitialise la zone de staging tout en laissant votre dossier de travail en l'état. L'historique sera perdu (les commits suivant <commit> seront perdus, mais pas vos modifications). Cette commande vous permet surtout de nettoyer l'historique en resoumettant un commit unique à la place de commit trop éparses.

git reset <commit de l'init par exemple>
git push -f origin master  # Force push master branch to github

avant je faisais ça, et je préfère
git checkout --orphan newbgit
	si fichiers à valider
	git add -A  # Add all files and commit them
	git commit
git branch -D master  # Deletes the master branch
git branch -m master  # Rename the current branch to master
git push -f origin master  # Force push master branch to github
```

### Debug
- go install github.com/go-delve/delve/cmd/dlv@latest
- avant de tester ouvrir la session SSH IONOS
