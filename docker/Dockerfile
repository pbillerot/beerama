# IMAGE beerama

# ETAPE COMPILATION DANS UNE IMAGE GOLANG ALPINE
# Le GOPATH par défaut de cette image est /go.
FROM golang:alpine AS goalpine

# Installation de GCC et GIT
RUN apk add build-base git

# Installation de beerama
WORKDIR /github/pbillerot
RUN git clone https://github.com/pbillerot/beerama.git

WORKDIR /github/pbillerot/beerama

# Compilation avec CGO
RUN GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /github/pbillerot/beerama/beerama

# ETAPE GENERATION D'UNE IMAGE ALPINE
# avec seulement l'exécutable golang
FROM alpine

# Ajout de tzdata pour régler l'heure
RUN apk add tzdata
ENV TZ=Europe/Paris
RUN cp /usr/share/zoneinfo/Europe/Paris /etc/localtime

# ajout de exiftool pour lire et modifier les metadata des images
RUN apk add exiftool

# Installation de beerama...
# seuls les répertoires conf et views ne seraient nécessaires
RUN mkdir -p /github/pbillerot/beerama
COPY --from=goalpine /github/pbillerot/beerama /github/pbillerot/beerama

# Le port sur lequel notre serveur écoute
EXPOSE 8080

# Point d'entrée
WORKDIR /github/pbillerot/beerama

ENTRYPOINT ["./beerama"]