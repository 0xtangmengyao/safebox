# safebox AI usage guide

[English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

This document is written to be **read or pasted by your own AI assistant**. It
gives that assistant enough context about `safebox` to explain the tool and to
draft commands, **without ever handling your secrets**.

## Read this first: secrets and the AI

> **Do not give the AI secrets, and do not let it ask you for secrets.**
>
> Never provide, paste, upload, or dictate any of the following to an AI
> assistant, a chatbot, or any online service:
>
> - Private keys or identity files (for example `AGE-SECRET-KEY-1...` secret keys, or the
>   contents of `./keys/alice.txt`).
> - Plaintext you want to protect.
> - Ciphertext (`*.age` files or armored `-----BEGIN AGE ENCRYPTED FILE-----`
>   blocks) that you consider sensitive.
> - Passphrases, API keys, tokens, or any other credential.
>
> The AI must **not** ask you for your private key, plaintext, ciphertext,
> passphrase, or tokens, and must not request that you display or paste the
> contents of any local key, input, or output file.

Rules the AI must follow when using this guide:

1. Base every answer only on this document and on **non-sensitive descriptions**
   you give it (for example "I have a private key at `./keys/alice.txt`" — the
   *path*, never the *contents*).
2. Only provide explanations and **draft** commands. Do not claim to have run
   anything.
3. **Do not execute commands.** Even if the AI has a shell or other tools, it must not use them for this task; it must only provide explanations and command drafts.
4. **Do not open, read, decrypt, encrypt, or upload any local key, input, or
   output file.**
5. Treat placeholders literally: `PATH`, `AGE1...`, `plain.txt`, `secret.age`
   are examples you must replace yourself.
6. The human operator must **review every command and run it personally**.

> This guide is **not** a technical control over AI systems. It cannot force an
> AI, a plugin, or a hosted service to behave or to forget data. You are
> responsible for following the privacy policy and data-handling rules of
> whichever AI service you choose, and for deciding what is safe to share with
> it.

## What safebox is

`safebox` is a small, safe, cross-platform command line tool written in Go for
encrypting local files and text. It also works well for moving ciphertext
between applications over a pipe.

- It is built only on the mature [age](https://age-encryption.org) standard
  library (`filippo.io/age`). It implements **no custom cryptography**.
- It uses the age **X25519** public-key scheme.
- Ciphertext is written as age **armor text**
  (`-----BEGIN AGE ENCRYPTED FILE-----`) by default. It can also read binary age
  input when decrypting.
- It reads and writes stdin/stdout by default and works with POSIX shells and
  PowerShell.
- **Plaintext is never accepted as a command line argument**, so it cannot leak
  into shell history or the process list.
- It is a purely local CLI: **no telemetry, no tracking, and no runtime network
  access of any kind**. It does not call any AI or HTTP API.

`safebox` has no `ask` command, no built-in AI, and no network features. Any AI
assistance comes from **your** assistant, outside this tool.

## Requirements

- Go **1.25 or newer** (the minimum version required by the current
  `filippo.io/age` dependency).
- No CGO required.

## Clone, build and install

```sh
# Clone
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox

# Build a local binary
go build -o safebox .

# Install into $GOBIN / %GOPATH%\bin
go install .
```

Cross-compile (pure Go, no CGO):

```sh
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

Check the build:

```sh
safebox version
safebox help
```

## Supported languages

The CLI help and diagnostics are available in:

| Code    | Language            |
| ------- | ------------------- |
| `en`    | English (default)   |
| `zh-CN` | Simplified Chinese  |
| `ja`    | Japanese            |
| `ko`    | Korean              |
| `es`    | Spanish             |
| `fr`    | French              |

Select a language with the global `--lang` flag (before the command) or the
`SAFEBOX_LANG` environment variable. `--lang` takes precedence; unknown values
fall back to English.

```sh
safebox --lang ja help
SAFEBOX_LANG=fr safebox decrypt --identity PATH
```

## Commands

All commands accept `-h` / `--help`.

### Generate a key pair (`keygen`)

```sh
safebox keygen --identity PATH
```

- Writes the private key to `PATH` (requesting mode `0600` on POSIX) and the
  public key to `PATH.pub` (requesting mode `0644`).
- On POSIX the effective permissions are also limited by your umask, so they are
  never wider than requested.
- On Windows, access is governed by the ACL inherited from the parent directory.
- **Never overwrites.** If either file already exists, it refuses to continue.
- Prints the public recipient (`age1...`) to stdout.

Example with an explicit placeholder path:

```sh
safebox keygen --identity ./keys/alice.txt
# private:   ./keys/alice.txt      (0600, expected)
# public:    ./keys/alice.txt.pub  (0644, expected)
# recipient: age1...               (also printed to stdout)
```

### Encrypt (`encrypt`)

```sh
safebox encrypt --recipient AGE1... [--recipient AGE1...] [--in PATH] [--out PATH]
```

- Reads plaintext from stdin unless `--in PATH` is given.
- Writes armor ciphertext to stdout unless `--out PATH` is given.
- `--recipient` may be repeated to encrypt to **multiple recipients**. Any one of
  the matching private keys can decrypt.
- Output files are created `0600` (POSIX, subject to umask) and are **never
  overwritten**.
- Plaintext must **not** be passed as an argument.

stdin/stdout example:

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1...
```

File example:

```sh
safebox encrypt --recipient AGE1... --in plain.txt --out secret.age
```

> Shell redirection such as `> file` is performed by the shell and may truncate an existing file before `safebox` starts. To save output safely, prefer `--out PATH`, which lets `safebox` refuse existing targets.

Multiple recipients:

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb --in plain.txt --out secret.age
```

### Decrypt (`decrypt`)

```sh
safebox decrypt --identity PATH [--in PATH] [--out PATH]
```

- Reads armored or binary age input from stdin unless `--in PATH` is given.
- Writes plaintext to stdout unless `--out PATH` is given.
- The identity file uses standard age rules: age-keygen style files are accepted,
  blank lines and `#` comment lines are ignored, and a single file may contain
  **multiple identities** (any matching identity can decrypt).
- Decryption first verifies the **complete** authentication tag in a
  permission-restricted temporary file; on failure no plaintext is emitted.
- If `--out` already exists, it refuses **before** decrypting or creating any
  temporary plaintext. Existing files are never overwritten or deleted.

stdin/stdout example:

```sh
safebox decrypt --identity PATH < secret.age
```

File example:

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

## Permissions and no-overwrite behavior (summary)

- `keygen`: private key `0600`, public key `0644` (POSIX requests, limited by
  umask). Neither file is overwritten.
- `encrypt`: output file `0600`, never overwritten.
- `decrypt`: output file `0600`, never overwritten; an existing `--out` target is
  rejected before any plaintext is produced. The existence check also rejects a
  dangling symlink.
- Ciphertext format is the age standard (armor text by default, binary on
  decrypt).

## What to ask your AI (safe examples)

- "Given that I have a private key at `./keys/alice.txt`, what exact `safebox`
  command encrypts `./report.pdf` to `./report.pdf.age` for the recipient
  `AGE1...`?" (Describe the *path*; never paste the key.)
- "Explain the difference between `--recipient` and `--identity`."
- "Draft a command that decrypts `./secret.age` to stdout without writing a
  file."
- "What permissions does `keygen` request, and why is that safe?"

## What NOT to ask your AI

- "Here is my private key, decrypt this for me."
- "Here is my plaintext/ciphertext, tell me what it says."
- "Run this command for me."
- "Open `./keys/alice.txt` and use it."

The AI should refuse these and instead explain the command you would run
yourself.
