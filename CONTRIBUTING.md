# Contributing to safebox

Thanks for your interest in improving `safebox`. This document explains how to
build, test and propose changes.

## Ground rules

- `safebox` uses the standard [age](https://age-encryption.org) implementation
  (`filippo.io/age`). **Do not add custom cryptographic algorithms or protocols.**
- **Do not change or weaken the existing behavior, flow, flags or security
  semantics** of `keygen`, `encrypt` and `decrypt`.
- The tool must stay a **local CLI with no runtime network access, telemetry or
  built-in AI**. Do not add HTTP/API clients or background requests.
- Never commit secrets, API keys, private keys or generated ciphertext.

## Requirements

- Go 1.25 or newer.
- No CGO required.

## Build and test

```sh
gofmt -l .        # list files that need formatting (should print nothing)
go vet ./...
go test ./...
go build ./...
```

All four must pass before you open a pull request. Continuous integration runs
the same checks and cross-compiles for Linux, macOS and Windows.

## Making changes

1. Fork the repository and create a topic branch.
2. Keep changes focused. Prefer small, reviewable commits.
3. Add or update tests for any behavior change.
4. Follow the existing code style; run `gofmt`.
5. Do not add third-party dependencies unless there is a strong, discussed
   reason. The project aims to stay on the Go standard library plus
   `filippo.io/age`.

## Localization

Translations live in `i18n_*.go` and the `README.<lang>.md` files. When you add
a message:

- Add the key to the authoritative English catalog (`i18n_en.go`) first.
- Add the same key to every other catalog. `go test ./...` fails if a
  translation is missing, empty, or changes the set of printf verbs.
- Keep terminology consistent: recipient, identity, plaintext, ciphertext,
  armor, temporary.

## Reporting issues

- Use the issue templates for bugs and feature requests.
- **Never include secrets.** Do not paste private keys, plaintext, ciphertext,
  API keys or real user data. Redact or synthesize examples.
- For security vulnerabilities, follow [SECURITY.md](SECURITY.md) instead of
  opening a public issue.

## Code of conduct

Participation is governed by the [Contributor Covenant](CODE_OF_CONDUCT.md).
