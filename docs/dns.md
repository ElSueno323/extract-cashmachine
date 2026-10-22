# Nom pour tout le réseau

Les postes du domaine utilisent le DNS `serveur-win.fonteynethekitchen.local` (`192.168.10.4`). Le nom `grandlivre.local` est une zone créée sur ce serveur. Il envoie les utilisateurs vers `192.168.10.210`, où nginx présente l’application.

`aperigroup.com` est le DNS public. Ce n’est pas là qu’il faut créer ce nom.

## Zone à créer

Dans le Gestionnaire DNS de `SERVEUR-WIN` :

1. Clic droit sur **Zones de recherche directes**, puis **Nouvelle zone**.
2. **Zone principale**.
3. Étendue de réplication : **Vers tous les serveurs DNS exécutés sur des contrôleurs de domaine dans ce domaine : fonteynethekitchen.local**.
4. **Zone de recherche directe**.
5. Nom de la zone : `grandlivre.local`.
6. Mises à jour dynamiques : ne pas les autoriser, ou n’autoriser que les mises à jour sécurisées.
7. Terminer.

La zone apparaît au même niveau que `fonteynethekitchen.local`. Elle ne se crée pas à l’intérieur de cette zone.

## Enregistrement

1. Clic droit sur `grandlivre.local`, puis **Nouvel hôte (A ou AAAA)**.
2. Laisser le nom vide. L’hôte est alors la zone elle-même : `grandlivre.local`.
3. Adresse IP : `192.168.10.210`.
4. Ajouter l’hôte.

PowerShell, exécuté sur `SERVEUR-WIN` avec un compte autorisé à écrire le DNS :

```powershell
Add-DnsServerPrimaryZone -Name "grandlivre.local" -ZoneFile "grandlivre.local.dns"
Add-DnsServerResourceRecordA -ZoneName "grandlivre.local" -Name "@" -IPv4Address "192.168.10.210"
```

Un poste du domaine n’a en général pas ce droit. La création se fait sur le contrôleur de domaine.

## Vérifier

```powershell
nslookup grandlivre.local 192.168.10.4
```

La réponse doit être `192.168.10.210`. Si un PC a cherché le nom avant la création de la zone, vider son cache :

```powershell
ipconfig /flushdns
```

Rien à modifier sur la machine Ubuntu : nginx accepte déjà ce nom et renvoie vers l’application.
