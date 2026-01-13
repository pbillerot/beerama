# merci à gemini
import os
import time

def nettoyer_vieux_fichiers(repertoire_racine, jours=30):
    # Conversion des jours en secondes
    secondes_limite = jours * 24 * 60 * 60
    temps_actuel = time.time()

    print(f"Nettoyage des fichiers de plus de {jours} jours dans : {repertoire_racine}")

    # Parcours de toute l'arborescence (fichiers et sous-dossiers)
    for dossier_parent, sous_dossiers, fichiers in os.walk(repertoire_racine):
        for nom_fichier in fichiers:
            chemin_complet = os.path.join(dossier_parent, nom_fichier)
            
            try:
                # Récupération de la date de dernière modification
                statut_fichier = os.stat(chemin_complet)
                age_fichier = temps_actuel - statut_fichier.st_mtime

                if age_fichier > secondes_limite:
                    print(f"Suppression : {chemin_complet}")
                    os.remove(chemin_complet)
                    
            except Exception as e:
                print(f"Erreur sur {chemin_complet} : {e}")

# --- CONFIGURATION ---
mon_dossier = "/volshare/data/photos/trash"
# go!
nettoyer_vieux_fichiers(mon_dossier, 30)