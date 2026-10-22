# Extract Cashmachine

Application web pour extraire les transactions des caisses ShopMaster et les télécharger en Excel. Le front est en Elm, le back en Go.

Sur le réseau, l’adresse est [http://grandlivre.local](http://grandlivre.local). Elle pointe vers le serveur `192.168.10.210`.

## Pour les utilisateurs

1. Ouvrir [http://grandlivre.local](http://grandlivre.local).
2. Cocher une ou plusieurs caisses.
3. Choisir la période.
4. Lancer l’extraction.

Chaque caisse qui répond produit un fichier Excel. S’il y en a plusieurs, un zip les regroupe. Une caisse injoignable est ignorée, les autres continuent. Le détail est dans [docs/utilisation.md](docs/utilisation.md).

Si le nom ne s’ouvre pas, lancer `ipconfig /flushdns` puis réessayer. Le nom est créé dans le DNS interne, voir [docs/dns.md](docs/dns.md).

## Documentation

| Document | Contenu |
| --- | --- |
| [docs/utilisation.md](docs/utilisation.md) | Écran, fichiers Excel, erreurs |
| [docs/dns.md](docs/dns.md) | Nom `grandlivre.local` pour tout le réseau |
| [docs/deploiement.md](docs/deploiement.md) | Service sur `192.168.10.210` |
| [docs/configuration.md](docs/configuration.md) | Fichier `.env` |
| [docs/developpement.md](docs/developpement.md) | Code, compilation, API |

## Sécurité

Le fichier `.env` n’est pas dans le dépôt. Il contient les adresses des caisses, le délai et le port d’écoute. L’interface ne reçoit que les noms des boutiques.

Ne pas y mettre de mot de passe, et ne pas committer `.env`.
