#!/bin/bash
# Récupère master et remplace le programme. Le fichier .env du serveur n'est pas touché.
set -euo pipefail

export PATH="/usr/local/go/bin:/usr/local/node/bin:${PATH}"
REPO=/opt/grand-livre/src
LOCK=/opt/grand-livre/deploy.lock

mkdir -p /opt/grand-livre
exec 9>"${LOCK}"
flock -n 9 || exit 0

if [ ! -d "${REPO}/.git" ]; then
  git clone --branch master --depth 1 https://github.com/ElSueno323/extract-cashmachine.git "${REPO}"
  changed=1
else
  git -C "${REPO}" fetch --depth 1 origin master
  local_rev="$(git -C "${REPO}" rev-parse HEAD)"
  remote_rev="$(git -C "${REPO}" rev-parse origin/master)"
  if [ "${local_rev}" = "${remote_rev}" ]; then
    exit 0
  fi
  git -C "${REPO}" reset --hard origin/master
  changed=1
fi

if [ "${changed}" != 1 ]; then
  exit 0
fi

cd "${REPO}/web"
npm ci
npx elm make src/Main.elm --output=../server/static/elm.js --optimize

cd "${REPO}"
CGO_ENABLED=0 go build -o /tmp/grand-livre-new ./server
sudo install -o ubuntu -g ubuntu -m 755 /tmp/grand-livre-new /opt/grand-livre/grand-livre
rm -f /tmp/grand-livre-new
sudo systemctl restart grand-livre
sleep 1
systemctl is-active grand-livre
curl -fsS http://127.0.0.1:8080/api/sites >/dev/null
