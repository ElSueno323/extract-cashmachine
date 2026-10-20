# Configuration

Toute la configuration sensible est dans `.env`, à la racine du projet ou à côté de `grand-livre.exe`. Le serveur le cherche d’abord dans le dossier courant, puis à côté de l’exécutable.

Copier `.env.example` vers `.env`, puis remplacer les valeurs. Le fichier réel ne doit pas être commité.

## Variables

| Variable | Rôle |
| --- | --- |
| `LISTEN_ADDR` | Adresse d’écoute de l’application, par exemple `127.0.0.1:8080`. |
| `HTTP_TIMEOUT_SECONDS` | Délai maximum, en secondes, pour joindre une caisse. Entier strictement positif. |
| `CSV_PATH` | Chemin du CSV sur chaque caisse. Il doit commencer par `/` et ne peut pas contenir `..` ni `://`. |
| `SITE_1`, `SITE_2`, … | Une caisse par ligne, sans trou dans la numérotation. La lecture s’arrête à la première ligne absente. |

## Format d’une caisse

```
SITE_1=identifiant|Nom affiché|adresse
```

- `identifiant` : minuscules, chiffres et tirets, 40 caractères au plus. Il sert dans le nom du fichier Excel.
- `Nom affiché` : libellé montré dans l’interface.
- `adresse` : adresse IP ou nom d’hôte de la caisse. Elle n’est jamais renvoyée au navigateur.

Exemple fictif :

```
LISTEN_ADDR=127.0.0.1:8080
HTTP_TIMEOUT_SECONDS=20
CSV_PATH=/ShopMasterTransactions2.csv
SITE_1=boutique-nord|Boutique Nord|10.0.0.40
SITE_2=boutique-sud|Boutique Sud|10.0.1.40
```

Deux lignes ne peuvent pas partager le même identifiant.

## Requête envoyée à une caisse

Pour une période du 11 septembre 2026 au 15 septembre 2026, le serveur appelle :

```
http://{adresse}{CSV_PATH}?startdate=11%2F09%2F2026&enddate=15%2F09%2F2026
```

Les dates sont au format `JJ/MM/AAAA`. Le CSV attendu utilise `;` comme séparateur.

Si `CSV_PATH`, `LISTEN_ADDR`, le délai ou `SITE_1` est absent ou invalide, le programme s’arrête et indique la clé en cause. Aucune adresse n’est écrite dans ce message d’erreur.
