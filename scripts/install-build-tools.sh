#!/bin/bash
set -euo pipefail
if [ ! -x /usr/local/go/bin/go ]; then
  curl -fsSL -o /tmp/go.tgz https://go.dev/dl/go1.27.3.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
fi
if [ ! -x /usr/local/node/bin/node ]; then
  ver=$(python3 -c 'import json,urllib.request; data=json.load(urllib.request.urlopen("https://nodejs.org/dist/index.json"));
print(next(item["version"] for item in data if item["version"].startswith("v22.")))')
  curl -fsSL -o /tmp/node.tar.xz "https://nodejs.org/dist/${ver}/node-${ver}-linux-x64.tar.xz"
  sudo rm -rf /usr/local/node
  sudo mkdir -p /usr/local/node
  sudo tar -C /usr/local/node --strip-components=1 -xJf /tmp/node.tar.xz
  rm -f /tmp/node.tar.xz
fi
/usr/local/go/bin/go version
/usr/local/node/bin/node --version
