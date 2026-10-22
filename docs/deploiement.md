# Déploiement

L’application tourne sur `192.168.10.210` (Ubuntu, nginx). Le programme écoute seulement sur `127.0.0.1:8080`. Nginx publie cette adresse sur le port 80, donc [http://grandlivre.local](http://grandlivre.local) et [http://192.168.10.210](http://192.168.10.210) affichent la même page.

## Emplacement

| Élément | Chemin |
| --- | --- |
| Programme | `/opt/grand-livre/grand-livre` |
| Configuration | `/opt/grand-livre/.env` |
| Service systemd | `grand-livre` |
| Site nginx | `/etc/nginx/sites-available/grand-livre` |
| Ancienne page d’accueil | `/etc/nginx/sites-available/default.bak-grand-livre` |

Le fichier `.env` appartient à l’utilisateur `ubuntu` et n’est lisible que par lui. Il reprend les mêmes clés que [configuration.md](configuration.md). `LISTEN_ADDR` reste `127.0.0.1:8080` pour que l’application ne soit pas exposée directement.

## Service

Le service démarre avec la machine et redémarre en cas d’arrêt.

```bash
sudo systemctl status grand-livre
sudo systemctl restart grand-livre
sudo journalctl -u grand-livre -n 50 --no-pager
```

Après un changement de `.env` :

```bash
sudo systemctl restart grand-livre
```

## Nginx

Le site actif est un reverse proxy :

```nginx
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;
    client_max_body_size 2m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 90s;
        proxy_send_timeout 90s;
    }
}
```

Le délai de 90 secondes couvre l’attente des caisses. Vérifier puis recharger :

```bash
sudo nginx -t && sudo systemctl reload nginx
```

## Mettre à jour le programme

Depuis un poste de développement, produire un binaire Linux :

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o grand-livre ./server
```

Le déposer à la place de `/opt/grand-livre/grand-livre`, puis :

```bash
sudo systemctl restart grand-livre
```

L’interface Elm est déjà incluse dans le binaire. Recompiler `web/src/Main.elm` vers `server/static/elm.js` avant `go build`, sinon le serveur embarque l’ancienne page.

## Contrôle rapide

Sur le serveur :

```bash
curl -fsS http://127.0.0.1:8080/api/sites
```

Depuis un poste du réseau :

```powershell
Invoke-WebRequest http://grandlivre.local/api/sites -UseBasicParsing
```

La réponse contient les noms des caisses, pas leurs adresses.
