# Utilisation

Sur le réseau, ouvrir [http://grandlivre.local](http://grandlivre.local). Cette adresse est le service installé sur `192.168.10.210`.

`grand-livre.exe` et `go run ./server` servent la même application en local, sur l’adresse de `LISTEN_ADDR`, en général `http://127.0.0.1:8080`.

La période proposée va du premier jour du mois en cours à aujourd’hui. Les caisses connues du `.env` sont cochées.

## Extraire

1. Décocher les boutiques à laisser de côté. Au moins une caisse doit rester cochée.
2. Choisir les dates de début et de fin. Le début doit être antérieur ou égal à la fin.
3. Lancer l’extraction.

Les caisses sont interrogées en parallèle. Le délai de chaque appel est `HTTP_TIMEOUT_SECONDS`.

## Résultat

Chaque caisse réussie affiche :

- le nombre de transactions (la ligne d’en-tête du CSV n’est pas comptée) ;
- un aperçu des 250 premières lignes ;
- un lien `Télécharger .xlsx`.

Le classeur s’appelle `{identifiant}_{début}_{fin}.xlsx`, par exemple `boutique-nord_11-09-2026_15-09-2026.xlsx`. La feuille `Transactions` reprend les colonnes du CSV. La première ligne est en gras et reste visible au défilement.

À partir de deux fichiers réussis, le lien `Tout télécharger (.zip)` regroupe les Excel de cette extraction.

Les liens de téléchargement restent valables deux heures, puis le serveur les oublie.

## Caisse ignorée

Une caisse est ignorée, sans interrompre les autres, lorsque :

- elle ne répond pas ou le délai est dépassé ;
- elle renvoie une erreur HTTP ;
- la réponse n’est pas un CSV (par exemple une page HTML).

L’écran affiche `Ignorée` et une raison courte. Cette raison ne contient pas l’adresse de la caisse.

Une caisse qui répond avec seulement la ligne d’en-tête reste réussie : le fichier Excel est créé et le texte indique qu’il n’y a aucune transaction sur la période.

## Ancien script

`export_transactions.py` lit le même `.env` et écrit les Excel dans le dossier courant, en ligne de commande. L’interface web est l’outil à utiliser.
