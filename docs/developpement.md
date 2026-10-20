# Développement

## Arborescence

```
server/main.go            serveur HTTP, fichiers statiques, ouverture du navigateur
server/static/            page, styles et elm.js servi au navigateur
web/src/Main.elm          interface
internal/config/          lecture de .env
internal/shop/            téléchargement et découpage du CSV (séparateur ;)
internal/sheet/           classeur Excel et archive zip
internal/httpapi/         routes JSON et téléchargements
export_transactions.py    ancien outil en ligne de commande, même .env
```

Au lancement depuis la racine du dépôt, les fichiers de `server/static/` sont lus sur le disque. Un `elm.js` recompilé est donc pris en compte sans reconstruire Go. L’exécutable embarque une copie de `server/static/` pour fonctionner hors de ce dossier.

## Compiler l’interface

Depuis `web/` :

```powershell
npm install
npx elm make src/Main.elm --output=../server/static/elm.js
```

Elm 0.19.1. Les dépendances sont déclarées dans `web/elm.json`.

## Compiler le serveur

Depuis la racine, avec Go 1.22 ou plus récent :

```powershell
go test ./...
go build -o grand-livre.exe ./server
```

`go test ./...` couvre le découpage CSV, le format des lignes de caisses, et le fait que `GET /api/sites` ne renvoie pas les adresses.

## API

Les dates de l’API sont en `AAAA-MM-JJ`. Le serveur les convertit en `JJ/MM/AAAA` pour les caisses.

### `GET /api/sites`

```json
{
  "defaultStart": "2026-10-01",
  "defaultEnd": "2026-10-20",
  "sites": [{ "id": "boutique-nord", "name": "Boutique Nord" }]
}
```

### `POST /api/export`

```json
{
  "sites": ["boutique-nord", "boutique-sud"],
  "start": "2026-09-11",
  "end": "2026-09-15"
}
```

Réponse `200` :

```json
{
  "results": [
    {
      "id": "boutique-nord",
      "name": "Boutique Nord",
      "ok": true,
      "transactions": 38,
      "filename": "boutique-nord_11-09-2026_15-09-2026.xlsx",
      "download": "/api/download/…",
      "headers": ["AID", "Start"],
      "rows": [["189658", "11/09/2026 11:43:33"]],
      "truncated": false,
      "error": ""
    }
  ],
  "zipDownload": "/api/download/…"
}
```

`rows` contient au plus 250 transactions. `truncated` vaut `true` quand l’aperçu est coupé. Le fichier Excel, lui, contient toutes les lignes. `zipDownload` est vide s’il y a moins de deux fichiers réussis.

Une caisse du `.env` qui échoue reste dans `results` avec `ok` à `false` et un `error` renseigné. Un identifiant absent du `.env`, une période invalide ou une liste de caisses vide répondent `400` :

```json
{ "error": "choisissez au moins une caisse" }
```

### `GET /api/download/{token}`

Renvoie le `.xlsx` ou le `.zip` créé par l’extraction. `404` si le jeton est inconnu ou expiré (deux heures).

Le corps d’une requête d’export est limité à 1 Mo. Le CSV lu sur une caisse est limité à 20 Mo.
