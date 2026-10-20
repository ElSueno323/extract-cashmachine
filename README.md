# Extract Cashmachine

Application web pour extraire les transactions des caisses ShopMaster et les télécharger en Excel.

Le front est en [Elm](web/src/Main.elm). Le back est en Go. Les adresses des caisses et les autres paramètres sensibles sont lus dans un fichier `.env` local, qui n’est pas versionné.

## Démarrage

1. Copier le modèle de configuration :

```powershell
copy .env.example .env
```

2. Renseigner `.env` : une ligne `SITE_n` par caisse, plus le chemin du CSV. Voir [docs/configuration.md](docs/configuration.md).

3. Lancer l’application depuis la racine du projet :

```powershell
.\grand-livre.exe
```

Le navigateur s’ouvre sur l’adresse indiquée par `LISTEN_ADDR` (par défaut `http://127.0.0.1:8080`).

Sans l’exécutable déjà compilé :

```powershell
go run ./server
```

## Utilisation

Cocher une ou plusieurs caisses, choisir la période, puis lancer l’extraction. Chaque caisse qui répond produit un fichier `.xlsx`. S’il y en a plusieurs, un `.zip` les regroupe. Une caisse injoignable est ignorée, les autres continuent.

Le détail des écrans, des fichiers et des erreurs est dans [docs/utilisation.md](docs/utilisation.md).

## Documentation

| Document | Contenu |
| --- | --- |
| [docs/utilisation.md](docs/utilisation.md) | Écran, fichiers Excel, cas d’erreur |
| [docs/configuration.md](docs/configuration.md) | Variables du fichier `.env` |
| [docs/developpement.md](docs/developpement.md) | Architecture, API, compilation, tests |

## Sécurité

Ne jamais committer `.env`. Le dépôt contient seulement `.env.example`, avec des valeurs fictives. L’interface ne reçoit pas les adresses des caisses : le serveur les utilise pour interroger le réseau, puis renvoie les noms et les transactions.
