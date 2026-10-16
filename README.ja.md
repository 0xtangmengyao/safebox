# safebox

[English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md)

AI 利用ガイド：[English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

[![CI](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml/badge.svg)](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](go.mod)

`safebox` は、ローカルのファイルやテキストを暗号化するための、小さく安全でクロスプラットフォームな Go 製コマンドラインツールです。アプリケーション間でパイプを通して暗号文をやり取りする用途にも適しています。

- 成熟した [age](https://age-encryption.org) 標準ライブラリ（`filippo.io/age`）のみを使用します。**独自の暗号アルゴリズムやプロトコルは実装しません。**
- 暗号文は既定で age の **armor テキスト形式**（`-----BEGIN AGE ENCRYPTED FILE-----`）で出力され、コピー＆ペーストやパイプに便利です。
- 既定では stdin/stdout を読み書きし、POSIX シェルと PowerShell の両方で動作します。**平文をコマンドライン引数として渡すことはできません**。シェル履歴やプロセス一覧への漏えいを防ぎます。
- 純粋にローカルな CLI です。**実行時のネットワークアクセスもテレメトリも一切ありません。** 自らインターネットへ接続することはありません。

## 特長

- `filippo.io/age` による標準的な age/X25519 の暗号化・復号。
- ストリーム指向で、大きなファイルをメモリに載せる必要がありません。
- 既存ファイルの上書きを拒否し、`O_EXCL` で競合を防ぎます。
- 平文を書き出す前に認証タグを完全に検証します。
- CLI ヘルプは英語・簡体字中国語・日本語・韓国語・スペイン語・フランス語に対応します。

## インストールとビルド

Go 1.25 以降が必要です（現在の `filippo.io/age` 依存の最小バージョンに一致）。

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

クロスコンパイル（純粋な Go、CGO 不要）：

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

`$GOBIN` / `%GOPATH%\bin` へインストール：

```sh
go install .
```

### ビルド済みバイナリ

`dist/` ディレクトリには Linux、macOS、Windows 向けの amd64・arm64 ビルドが公開されています。お使いの OS と CPU アーキテクチャに合ったファイルを選んでください。`dist/SHA256SUMS` はダウンロードが壊れていないことの確認に使えますが、チェックサムだけでは公開者を証明できません。現時点でバイナリはコード署名されていません。機密性が高い用途ではソースからビルドし、出所を自分で確認してください。

バージョンとヘルプの表示：

```sh
safebox version
safebox help
```

## コマンド

### 鍵ペアの生成

```sh
safebox keygen --identity ./keys/alice.txt
```

- 秘密鍵は `./keys/alice.txt` に書き込まれ、POSIX システムでは権限 `0600` を要求します。公開鍵は `./keys/alice.txt.pub` に書き込まれ、権限 `0644` を要求します。
- POSIX システムでは**実効権限はプロセスの umask によっても制限されます**。umask は権限ビットを削除するだけなので、要求より広くなることはありません。たとえば umask `022` では公開鍵は `0644`、umask `077` では `0600` になります。秘密鍵は `0600` を要求し、グループやその他のビットを含まないため通常 umask の影響を受けません。
- Windows では、親ディレクトリから継承される ACL によってアクセスが制御されます。そのディレクトリが想定したアカウントだけにアクセス可能であることを確認してください。
- どちらかのファイルが既に存在する場合、`safebox` は処理を拒否します。**上書きは決して行いません。**
- 公開鍵（受信者）は stdout にも出力され、配布しやすくなっています。

### 暗号化

既定では stdin から平文を読み取り、armor 暗号文を stdout に出力します：

```sh
safebox encrypt --recipient age1... < plain.txt
```

受信者は複数指定できます（いずれか一致する秘密鍵で復号可能）：

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt
```

ファイルパスを使うこともできます。出力ファイルは権限 `0600` で作成され、**上書きされません**：

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

POSIX システムでの要求出力権限は `0600` です（同様に umask により狭くなることはあっても広くなることはありません）。Windows では継承された親ディレクトリの ACL が適用されます。

> shell の `> file` リダイレクトは shell が処理するため、`safebox` の起動前に既存ファイルを切り詰める場合があります。保存時は `--out PATH` を使うと、`safebox` の上書き拒否保護が適用されます。

### 復号

既定では stdin から暗号文を読み取り、平文を stdout に出力します：

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age
```

ファイル形式（POSIX の出力権限 `0600`、Windows は継承 ACL、**上書きしません**）：

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

アイデンティティファイルは age 標準の規則で解析されます。age-keygen 形式のファイルに対応し、空行と `#` コメント行（`# created: ...` や `# public key: age1...` など）は無視され、1 つのファイルに**複数のアイデンティティ**を含めることができ、一致するものがあれば復号できます。

復号では、まず平文をシステムの一時ディレクトリ内の権限を制限した一時ファイルにストリームとして退避し、認証タグを完全に検証してから初めて書き出します。認証に失敗した場合、平文は一切出力されず、メモリ使用量がファイルサイズに比例して増えることもありません。一時ファイルは終了時に削除されますが、**削除は安全な消去ではありません**。一時ディレクトリのあるディスク、スワップ、バックアップにはデータが残る可能性があります。入力は armor テキストとバイナリ age の両方に対応します。

`--out` が既存のパスを指す場合、**復号や一時平文の作成の前に**拒否されます。この存在チェックはシンボリックリンクをたどらないため、壊れたシンボリックリンクも既存として拒否されます。最終的な書き出しでは引き続き `O_EXCL` を競合対策の最終防壁として使用します。既存の `--out` ターゲットが上書き・削除されることはありません。ただし、今回の実行で新規ファイルを作成した後に書き込みエラーが発生した場合は、今回生じた部分的な出力が削除されます。

## 使用例

### テキストのパイプ（Linux / macOS / Git Bash）

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1... | safebox decrypt --identity alice.txt
```

### ファイル

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient "$(cat alice.txt.pub)" --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### アプリケーション間の受け渡し

```sh
# アプリケーション A が暗号文をパイプへ出力
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# 鍵ペアを生成
safebox keygen --identity .\keys\alice.txt

# ファイルモードなら PowerShell によるエンコーディングや改行の書き換えを避けられる
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# 復号
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> 注：ファイルモードは PowerShell のテキストオブジェクトパイプラインを通らないため、任意のファイルに適しています。`echo`/`printf` コマンドの内容として機密テキストを直接入力すると、シェル履歴に残る可能性があります。機密テキストはファイル、クリップボードツール、または上流アプリから stdin 経由で渡してください。平文や秘密鍵を `safebox` のコマンドラインに書かないでください。

## 言語とローカライズ

CLI のヘルプと診断は、英語（`en`）、簡体字中国語（`zh-CN`）、日本語（`ja`）、韓国語（`ko`）、スペイン語（`es`）、フランス語（`fr`）に対応しています。既定は英語です。

グローバルな `--lang` フラグまたは `SAFEBOX_LANG` 環境変数で明示的に切り替えられます：

```sh
safebox --lang ja help      # 日本語のヘルプ
SAFEBOX_LANG=fr safebox decrypt   # フランス語の診断
```

`--lang` は `SAFEBOX_LANG` より優先され、不明な値は英語にフォールバックします。自分の AI アシスタントに `safebox` を安全に使わせる方法は [AI 利用ガイド](AI_GUIDE.md)（[简体中文](AI_GUIDE.zh-CN.md)）を参照してください。

## セキュリティ境界

- **暗号技術**：`filippo.io/age` による標準の age / X25519。`safebox` は独自の暗号を発明しません。
- **秘密鍵の保護**：POSIX システムでは秘密鍵ファイルは `0600` で作成され、Windows では親ディレクトリの ACL に従います。`safebox` は秘密鍵を決して表示しません。鍵ファイルは自分で適切に保護し、共用または非暗号化のディスクに置かないでください。
- **一時平文**：復号ではストリーム全体を検証するために権限を制限した一時ファイルを使用します。終了後に削除されますが、SSD、スワップ、バックアップ、スナップショットから安全に消去される保証はありません。高い機密性が必要な場合は、一時ディレクトリとスワップを信頼できる暗号化ストレージに置いてください。
- **復元不可**：秘密鍵を失うと、暗号文は**復元できません**。バックドアも復元機構もありません。
- **メタデータ**：暗号化が保護するのはファイルの内容だけです。age はファイルサイズ、タイムスタンプ、ファイル名を隠さず、通信パターンも隠しません。
- **端末とクリップボード**：端末経由で鍵や平文をコピー＆ペーストすると、端末履歴、スクロールバック、クリップボードに残る可能性があります。
- **パスフレーズのエントロピー**：本ツールは公開鍵暗号を使用し、パスフレーズに依存しません。パスフレーズ暗号化には低エントロピーのパスフレーズを使わないでください。
- **ローカルツール**：`safebox` はローカル CLI で、実行時のネットワーク機能もテレメトリも一切ありません。

## 開発とテスト

```sh
gofmt -l .
go vet ./...
go test ./...
```

テストは、鍵生成とファイル権限（プラットフォーム別に調整）、stdin/stdout とファイルの往復、複数受信者、暗号文の改ざんや誤った鍵での失敗時に平文が漏れないこと、既存ファイルの上書き拒否を網羅します。さらに、出荷されるすべての言語の翻訳の完全性も検証します。

継続的インテグレーションは、すべての push と pull request でビルドとテストを実行し、`gofmt`、`go vet`、`go test` を走らせ、Linux・macOS・Windows 向けにクロスコンパイルします（成果物はリポジトリに残しません）。

## プロジェクトメタデータ（提案）

以下の GitHub topics は提案です。リポジトリ設定で任意に設定してください。本リポジトリはオンラインのメタデータを自動変更しません：

`age`、`encryption`、`cryptography`、`x25519`、`cli`、`golang`、`privacy`、`command-line-tool`、`file-encryption`、`cross-platform`

## コントリビューション

ビルド、テスト、変更の提出方法は [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。脆弱性は [SECURITY.md](SECURITY.md) の方法で非公開で報告してください。本プロジェクトは [Contributor Covenant](CODE_OF_CONDUCT.md) に従います。

## ライセンス

MIT。詳しくは [LICENSE](LICENSE) を参照してください。
