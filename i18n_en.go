package main

// enMessages is the authoritative English catalog. Every other catalog must
// define exactly the same keys (enforced by TestCatalogsCoverAllLanguages).
var enMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - age-based file and text encryption",
	"usage.header":      "Usage:",
	"usage.synopsis":    "  safebox [--lang CODE] <command> [flags]",
	"usage.commands":    "Commands:",
	"usage.keygen":      "  keygen   Generate an X25519 identity and its public recipient",
	"usage.encrypt":     "  encrypt  Encrypt stdin (or --in) to armor text on stdout (or --out)",
	"usage.decrypt":     "  decrypt  Decrypt armor/binary age input to stdout (or --out)",
	"usage.version":     "  version  Print the version",
	"usage.help":        "  help     Print this help",
	"usage.commandHelp": `Run "safebox <command> -h" for command specific flags.`,
	"usage.plaintext": "Plaintext is never accepted as a command line argument: pipe it in via stdin\n" +
		"  to avoid leaking it to shell history and the process list.",
	"usage.lang": "Language: use --lang en|zh-CN|ja|ko|es|fr or the SAFEBOX_LANG environment variable.",

	// Diagnostics (printed as "safebox: <text>").
	"diag.missingCommand":            "missing command",
	"diag.unknownCommand":            "unknown command %q",
	"diag.unexpectedArgument":        "%s: unexpected argument %q",
	"diag.keygen.requiresIdentity":   "keygen requires --identity PATH",
	"diag.encrypt.requiresRecipient": "encrypt requires at least one --recipient",
	"diag.decrypt.requiresIdentity":  "decrypt requires --identity PATH",
	"diag.interrupted":               "interrupted",

	// Per-command usage.
	"keygen.usageLine": "Usage: safebox keygen --identity PATH",
	"keygen.detail": "Writes PATH (private key, 0600) and PATH.pub (public key, 0644).\n" +
		"Neither file is ever overwritten.",

	"encrypt.usageLine": "Usage: safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "Reads plaintext from stdin unless --in is given. Writes armored age text.\n" +
		"Plaintext must not be passed as an argument.",

	"decrypt.usageLine": "Usage: safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "Accepts armored or binary age input. Plaintext is verified in full\n" +
		"before anything is written, so a failure never leaks partial output.",

	// Flag descriptions (shown by flag.PrintDefaults).
	"flag.keygen.identity":   "path of the X25519 private key file to create (0600)",
	"flag.encrypt.recipient": "recipient public key (age1...); may be repeated",
	"flag.encrypt.in":        "input file (default stdin)",
	"flag.encrypt.out":       "output file (default stdout); never overwritten",
	"flag.decrypt.identity":  "path of the X25519 private key file",
	"flag.decrypt.in":        "input file (default stdin)",
	"flag.decrypt.out":       "output file (default stdout); never overwritten",
}
