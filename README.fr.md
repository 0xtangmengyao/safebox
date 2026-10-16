# safebox

[English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md)

Guide d'usage de l'IA : [English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

[![CI](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml/badge.svg)](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](go.mod)

`safebox` est un petit outil en ligne de commande Go, sûr et multiplateforme, pour chiffrer des fichiers et du texte locaux. Il convient aussi au transfert de texte chiffré entre applications via un tube.

- Repose uniquement sur la bibliothèque standard mature [age](https://age-encryption.org) (`filippo.io/age`). **Aucun algorithme ni protocole maison.**
- Le texte chiffré est écrit par défaut au format **texte armor** d'age (`-----BEGIN AGE ENCRYPTED FILE-----`), facile à copier, coller et rediriger.
- Lit et écrit stdin/stdout par défaut et fonctionne aussi bien avec les shells POSIX que PowerShell. **Le texte en clair n'est jamais accepté comme argument de ligne de commande**, il ne peut donc pas fuiter dans l'historique du shell ni la liste des processus.
- Une CLI purement locale : **aucune télémétrie, aucun suivi et aucun accès réseau d'aucune sorte en cours d'exécution.** Elle ne se connecte jamais à Internet de sa propre initiative.

## Points forts

- Chiffrement et déchiffrement age/X25519 standard via `filippo.io/age`.
- Orienté flux : les gros fichiers n'ont pas besoin de tenir en mémoire.
- Refuse d'écraser un fichier existant et utilise `O_EXCL` pour éviter les conditions de course.
- Vérifie intégralement l'étiquette d'authentification avant d'écrire le moindre texte en clair.
- Aide de la CLI localisée en anglais, chinois simplifié, japonais, coréen, espagnol et français.

## Installation et compilation

Go 1.25 ou supérieur est requis (la version minimale de la dépendance actuelle `filippo.io/age`).

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

Compilation croisée (Go pur, sans CGO) :

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

Installer dans `$GOBIN` / `%GOPATH%\bin` :

```sh
go install .
```

### Binaires précompilés

Le répertoire `dist/` publie des builds Linux, macOS et Windows pour amd64 et arm64. Choisissez le fichier correspondant à votre système d'exploitation et à votre architecture de processeur. `dist/SHA256SUMS` permet de vérifier qu'un téléchargement est intact, mais une somme de contrôle seule ne prouve pas qui l'a publié. Les binaires ne sont pas encore signés ; pour un usage très sensible, compilez depuis les sources et vérifiez-en vous-même l'origine.

Afficher la version et l'aide :

```sh
safebox version
safebox help
```

## Commandes

### Générer une paire de clés

```sh
safebox keygen --identity ./keys/alice.txt
```

- La clé privée est écrite dans `./keys/alice.txt` en demandant le mode `0600` sur les systèmes POSIX ; la clé publique est écrite dans `./keys/alice.txt.pub` en demandant le mode `0644`.
- Sur les systèmes POSIX, les **permissions effectives sont aussi limitées par l'umask du processus**. L'umask ne fait que retirer des bits de permission, le résultat n'est donc jamais plus large que demandé. Par exemple, avec umask `022` la clé publique est `0644`, et avec umask `077` elle est `0600`. La clé privée demande `0600`, qui n'a aucun bit de groupe ni autre, elle n'est donc normalement pas affectée par l'umask.
- Sous Windows, l'accès aux fichiers est contrôlé par l'ACL héritée du répertoire parent ; assurez-vous que ce répertoire n'est accessible qu'aux comptes prévus.
- Si l'un des fichiers existe déjà, `safebox` refuse de continuer. Il **n'écrase jamais**.
- La clé publique (destinataire) est aussi imprimée sur stdout pour la partager facilement.

### Chiffrer

Par défaut, le texte en clair est lu depuis stdin et le texte chiffré armor est écrit sur stdout :

```sh
safebox encrypt --recipient age1... < plain.txt
```

Vous pouvez indiquer plusieurs destinataires ; toute clé privée correspondante peut déchiffrer :

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt
```

Vous pouvez aussi utiliser des chemins de fichiers. Les fichiers de sortie sont créés en mode `0600` et ne sont **jamais écrasés** :

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

Sur les systèmes POSIX, le mode de sortie demandé est `0600` (de nouveau limité par l'umask, uniquement plus étroit, jamais plus large). Sous Windows, l'ACL héritée du répertoire parent s'applique.

> Une redirection shell telle que `> fichier` est gérée par le shell et peut tronquer un fichier existant avant le démarrage de `safebox`. Pour enregistrer la sortie, utilisez `--out PATH` afin de conserver la protection anti-écrasement de `safebox`.

### Déchiffrer

Par défaut, le texte chiffré est lu depuis stdin et le texte en clair est écrit sur stdout :

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age
```

Forme avec fichiers (mode de sortie POSIX `0600`, ACL héritée sous Windows, **n'écrase jamais**) :

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

Les fichiers d'identité sont analysés selon les règles standard d'age : les fichiers de style age-keygen sont acceptés, les lignes vides et les lignes de commentaire `#` (comme `# created: ...` et `# public key: age1...`) sont ignorées, et un même fichier peut contenir **plusieurs identités** ; toute identité correspondante peut déchiffrer.

Le déchiffrement place d'abord le texte en clair dans un fichier temporaire à permissions restreintes dans le répertoire temporaire du système, et ne l'écrit qu'après vérification complète de l'étiquette d'authentification. En cas d'échec d'authentification, aucun texte en clair n'est émis, et l'usage mémoire n'augmente pas avec la taille du fichier. Le fichier temporaire est supprimé à la sortie, mais **supprimer n'est pas un effacement sûr** : le disque hébergeant le répertoire temporaire, le swap ou les sauvegardes peuvent encore retenir des données. Les entrées texte armor et age binaire sont toutes deux acceptées.

Si `--out` pointe vers un chemin existant, `safebox` refuse avant tout déchiffrement ou création de texte en clair temporaire. Cette vérification d'existence ne suit pas les liens symboliques, donc un lien mort compte aussi comme existant et est rejeté ; l'écriture finale utilise toujours `O_EXCL` comme garde-fou de référence. Une cible `--out` existante n'est jamais écrasée ni supprimée. Si cette exécution a créé un nouveau fichier puis rencontré un échec d'écriture, la sortie partielle de cette exécution est nettoyée.

## Exemples

### Texte via un tube (Linux / macOS / Git Bash)

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1... | safebox decrypt --identity alice.txt
```

### Fichiers

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient "$(cat alice.txt.pub)" --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### Transférer du texte chiffré entre applications

```sh
# L'application A produit du texte chiffré dans le tube
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# Générer une paire de clés
safebox keygen --identity .\keys\alice.txt

# Le mode fichier évite que PowerShell réécrive l'encodage ou les retours à la ligne
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# Déchiffrer
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> Remarque : le mode fichier ne passe pas par le pipeline d'objets texte de PowerShell, il convient donc aux fichiers arbitraires. Si vous saisissez du texte sensible directement dans une commande `echo`/`printf`, l'historique du shell peut l'enregistrer. Fournissez le texte sensible depuis un fichier, un outil de presse-papiers ou une application en amont via stdin. Ne mettez jamais de texte en clair ni de clé privée sur la ligne de commande de `safebox`.

## Langue et localisation

L'aide et les diagnostics de la CLI sont disponibles en anglais (`en`), chinois simplifié (`zh-CN`), japonais (`ja`), coréen (`ko`), espagnol (`es`) et français (`fr`). La valeur par défaut est l'anglais.

Changez de langue explicitement avec l'option globale `--lang` ou la variable d'environnement `SAFEBOX_LANG` :

```sh
safebox --lang ja help      # aide en japonais
SAFEBOX_LANG=fr safebox decrypt   # diagnostics en français
```

`--lang` est prioritaire sur `SAFEBOX_LANG`, et les valeurs inconnues reviennent à l'anglais. Pour savoir comment utiliser votre propre assistant IA avec `safebox`, consultez le [guide d'usage de l'IA](AI_GUIDE.md) ([简体中文](AI_GUIDE.zh-CN.md)).

## Limites de sécurité

- **Cryptographie** : age / X25519 standard via `filippo.io/age`. `safebox` n'invente aucune cryptographie.
- **Clés privées** : sur les systèmes POSIX, le fichier de clé privée est créé avec `0600` ; sous Windows, il est régi par l'ACL du répertoire parent. `safebox` n'imprime jamais la clé privée. Protégez vous-même le fichier de clé et ne le stockez pas sur un disque partagé ou non chiffré.
- **Texte en clair temporaire** : le déchiffrement utilise un fichier temporaire à permissions restreintes pour vérifier tout le flux. Il est supprimé ensuite, mais son effacement sûr depuis les SSD, le swap, les sauvegardes ou les instantanés n'est pas garanti. Pour une confidentialité élevée, placez le répertoire temporaire et le swap sur un stockage chiffré de confiance.
- **Aucune récupération** : si vous perdez une clé privée, le texte chiffré **ne peut pas être récupéré**. Il n'y a ni porte dérobée ni mécanisme de récupération.
- **Métadonnées** : le chiffrement ne protège que le contenu des fichiers. age ne masque ni la taille, ni l'horodatage, ni le nom des fichiers, et ne masque pas les motifs de trafic.
- **Terminaux et presse-papiers** : copier des clés ou du texte en clair via un terminal peut les laisser dans l'historique, le tampon de défilement ou le presse-papiers.
- **Entropie des phrases de passe** : cet outil utilise le chiffrement à clé publique et ne dépend pas des phrases de passe ; n'utilisez pas de phrases de faible entropie pour le chiffrement par phrase de passe.
- **Outil local** : `safebox` est une CLI locale, sans aucune fonction réseau ni télémétrie en cours d'exécution.

## Développement et tests

```sh
gofmt -l .
go vet ./...
go test ./...
```

La suite de tests couvre la génération de clés et les permissions de fichiers (adaptées par plateforme), les allers-retours stdin/stdout et par fichier, les destinataires multiples, l'échec sans fuite de texte en clair lorsque le texte chiffré est altéré ou la mauvaise clé utilisée, et le refus d'écraser des fichiers existants. Elle vérifie aussi l'exhaustivité des traductions pour chaque langue publiée.

L'intégration continue compile et teste chaque push et pull request, exécute `gofmt`, `go vet` et `go test`, et compile en croisé pour Linux, macOS et Windows sans laisser d'artefacts dans le dépôt.

## Métadonnées du projet (suggérées)

Ces topics GitHub sont des suggestions que vous pouvez définir dans les paramètres du dépôt ; ce dépôt ne modifie pas automatiquement les métadonnées en ligne :

`age`, `encryption`, `cryptography`, `x25519`, `cli`, `golang`, `privacy`, `command-line-tool`, `file-encryption`, `cross-platform`

## Contribuer

Voir [CONTRIBUTING.md](CONTRIBUTING.md) pour compiler, tester et proposer des modifications. Signalez les vulnérabilités en privé comme décrit dans [SECURITY.md](SECURITY.md). Ce projet suit le [Contributor Covenant](CODE_OF_CONDUCT.md).

## Licence

MIT. Voir [LICENSE](LICENSE).
