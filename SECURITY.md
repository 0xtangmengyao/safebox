# Security Policy

## Supported versions

`safebox` is pre-1.0. Security fixes are applied to the latest release on the
`main` branch.

## Reporting a vulnerability

Please report security issues **privately**. Do not open a public issue for a
suspected vulnerability.

- Preferred: open a private report through GitHub's
  [Security Advisories](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
  for this repository ("Report a vulnerability" on the Security tab).
- If you cannot use GitHub, contact the maintainers through their GitHub
  profile. Do not post details publicly before a fix is available.

Please include:

- A description of the issue and its impact.
- Steps to reproduce, ideally with a minimal example.
- Affected version and platform.
- Any suggested remediation.

We aim to acknowledge reports promptly and to coordinate disclosure after a fix
is available.

## Please do not send secrets

Never include real private keys, plaintext, ciphertext, API keys, passphrases or
personal data in a report. Use synthetic or redacted examples.

## Scope

In scope:

- The `safebox` CLI (`main.go`, `i18n*.go`).
- Handling of key files, permissions, temporary plaintext and output files.

Out of scope:

- Vulnerabilities in `filippo.io/age` or the Go standard library (report those
  upstream).
- Issues that require an already-compromised machine or a malicious local user
  with access to the user's files.
- The inherent limitations documented in the README, such as the inability to
  securely erase temporary files from SSDs, swap or backups.
