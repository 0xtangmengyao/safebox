# safebox

[English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md)

AI usage guide: [English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

[![CI](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml/badge.svg)](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](go.mod)

`safebox` is a small, safe, cross-platform Go command line tool for encrypting
local files and text. It also works well for moving ciphertext between
applications over a pipe.

- Built only on the mature [age](https://age-encryption.org) standard library
  (`filippo.io/age`). **No custom ciphers or protocols.**
- Ciphertext is written as age **armor text**
  (`-----BEGIN AGE ENCRYPTED FILE-----`) by default, so it is easy to copy,
  paste and pipe.
- Reads and writes stdin/stdout by default and works with both POSIX shells and
  PowerShell. **Plaintext is never accepted as a command line argument**, so it
  cannot leak into shell history or the process list.
- A purely local CLI: **no telemetry, no tracking, and no runtime network
  access of any kind.** It never connects to the internet on its own.

## Highlights

- Standard age/X25519 encryption and decryption via `filippo.io/age`.
- Stream-oriented, so large files do not need to fit in memory.
- Refuses to overwrite existing files, using `O_EXCL` to avoid races.
- Verifies the full authentication tag before writing any plaintext.
- Localized CLI help in English, Simplified Chinese, Japanese, Korean, Spanish
  and French.

## Install and build

Go 1.25 or newer is required (the minimum version of the current
`filippo.io/age` dependency).

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

Cross-compilation (pure Go, no CGO required):

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

Install into `$GOBIN` / `%GOPATH%\bin`:

```sh
go install .
```

### Prebuilt binaries

The `dist/` directory publishes Linux, macOS and Windows builds for amd64 and
arm64. Pick the file for your operating system and CPU architecture.
`dist/SHA256SUMS` lets you check that a download is intact, but a checksum alone
does not prove who published it. The binaries are not code signed yet; for
highly sensitive use, build from source and verify the source yourself.

Print the version and help:

```sh
safebox version
safebox help
```

## Commands

### Generate a key pair

```sh
safebox keygen --identity ./keys/alice.txt
```

- The private key is written to `./keys/alice.txt` requesting mode `0600` on
  POSIX systems; the public key is written to `./keys/alice.txt.pub` requesting
  mode `0644`.
- On POSIX systems the **effective permissions are also limited by your process
  umask**. umask only removes permission bits, so the result is never wider than
  requested. For example, with umask `022` the public key is `0644`, and with
  umask `077` it is `0600`. The private key requests `0600`, which has no group
  or other bits, so it is normally unaffected by umask.
- On Windows, file access is controlled by the ACL inherited from the parent
  directory; make sure that directory is only accessible to the intended
  accounts.
- If either file already exists, `safebox` refuses to continue. It **never
  overwrites**.
- The public key (recipient) is also printed to stdout so it is easy to share.

### Encrypt

By default, plaintext is read from stdin and armor ciphertext is written to
stdout:

```sh
safebox encrypt --recipient age1... < plain.txt
```

You can pass several recipients; any matching private key can decrypt:

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt
```

You can also use file paths. Output files are created with mode `0600` and are
**never overwritten**:

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

On POSIX systems the requested output mode is `0600` (again limited by umask,
only narrower, never wider). On Windows, the inherited parent-directory ACL
applies.

> Shell redirection such as `> file` is handled by your shell and may truncate an existing file before `safebox` starts. Use `--out PATH` to save through `safebox` and retain its no-overwrite protection.

### Decrypt

By default, ciphertext is read from stdin and plaintext is written to stdout:

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age
```

File form (POSIX output mode `0600`, Windows inherited ACL, **never
overwrites**):

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

Identity files are parsed with the standard age rules: age-keygen style files
are accepted, blank lines and `#` comment lines (such as `# created: ...` and
`# public key: age1...`) are ignored, and a single file may contain **multiple
identities**; any matching identity can decrypt.

Decryption first spools plaintext into a permission-restricted temporary file in
the system temporary directory and only writes it out after the complete
authentication tag has been verified. On authentication failure, no plaintext is
emitted, and memory use does not grow with the file size. The temporary file is
removed on exit, but **removal is not secure erasure**: the disk holding the
temporary directory, swap, or backups may still retain data. Both armor text and
binary age input are accepted.

If `--out` points at an existing path, `safebox` refuses before decrypting or
creating any temporary plaintext. That existence check does not follow symbolic
links, so a dangling symlink also counts as existing and is rejected; the final
write still uses `O_EXCL` as the authoritative guard. An existing `--out` target
is never overwritten or deleted. If this run created a new file and then hit a
write failure, the partial output from this run is cleaned up.

## Examples

### Text over a pipe (Linux / macOS / Git Bash)

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1... | safebox decrypt --identity alice.txt
```

### Files

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient AGE1... --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### Moving ciphertext between applications

```sh
# Application A produces ciphertext into the pipe
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# Generate a key pair
safebox keygen --identity .\keys\alice.txt

# Use file mode to avoid PowerShell rewriting text encoding or newlines
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# Decrypt
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> Note: file mode does not pass through PowerShell's text object pipeline, so it
> is suitable for arbitrary files. If you type sensitive text directly into an
> `echo`/`printf` command, the shell history may still record it. Feed sensitive
> text from a file, a clipboard tool, or an upstream application via stdin. Never
> put plaintext or a private key on the `safebox` command line.

## Language and localization

The CLI help and diagnostics are available in English (`en`), Simplified
Chinese (`zh-CN`), Japanese (`ja`), Korean (`ko`), Spanish (`es`) and French
(`fr`). The default is English.

Switch language explicitly with the global `--lang` flag or the `SAFEBOX_LANG`
environment variable:

```sh
safebox --lang ja help      # Japanese help
SAFEBOX_LANG=fr safebox decrypt   # French diagnostics
```

`--lang` takes precedence over `SAFEBOX_LANG`, and unknown values fall back to
English. For guidance on using your own AI assistant to help with `safebox`,
see the [AI usage guide](AI_GUIDE.md) ([简体中文](AI_GUIDE.zh-CN.md)).

## Security boundaries

- **Cryptography**: standard age / X25519 via `filippo.io/age`. `safebox` invents
  no cryptography of its own.
- **Private keys**: on POSIX systems the private key file is created with `0600`;
  on Windows it is governed by the parent-directory ACL. `safebox` never prints
  the private key. Protect the key file yourself and do not store it on shared or
  unencrypted disks.
- **Temporary plaintext**: decryption uses a permission-restricted temporary file
  to verify the whole stream. It is deleted afterwards, but is not guaranteed to
  be securely erased from SSDs, swap, backups or snapshots. For high
  confidentiality, keep the temporary directory and swap on trusted encrypted
  storage.
- **No recovery**: if you lose a private key, the ciphertext **cannot be
  recovered**. There is no backdoor and no recovery mechanism.
- **Metadata**: encryption protects only file contents. age does not hide file
  size, timestamps or file names, and does not hide traffic patterns.
- **Terminals and clipboards**: copying keys or plaintext through a terminal can
  leave them in terminal history, scrollback or the clipboard.
- **Passphrase entropy**: this tool uses public-key encryption and does not rely
  on passphrases; do not use low-entropy passphrases for passphrase encryption.
- **Local tool**: `safebox` is a local CLI with no runtime network or telemetry
  features of any kind.

## Development and testing

```sh
gofmt -l .
go vet ./...
go test ./...
```

The test suite covers key generation and file permissions (adapted per
platform), stdin/stdout and file round trips, multiple recipients, failure
without plaintext leakage when ciphertext is tampered with or the wrong key is
used, and refusal to overwrite existing files. It also checks translation
completeness across every shipped language.

Continuous integration builds and tests every push and pull request, runs
`gofmt`, `go vet` and `go test`, and cross-compiles for Linux, macOS and Windows
without leaving artifacts in the repository.

## Project metadata (suggested)

These GitHub topics are suggestions you can set in the repository settings; this
repository does not change online metadata automatically:

`age`, `encryption`, `cryptography`, `x25519`, `cli`, `golang`, `privacy`,
`command-line-tool`, `file-encryption`, `cross-platform`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to build, test and submit
changes. Please report vulnerabilities privately as described in
[SECURITY.md](SECURITY.md). This project follows the
[Contributor Covenant](CODE_OF_CONDUCT.md).

## License

MIT. See [LICENSE](LICENSE).
