// Command safebox is a small, safe, cross-platform CLI for encrypting local
// files and text with the age file encryption format (filippo.io/age).
//
// It intentionally does not implement any cryptography of its own: all keys,
// recipients, encryption, decryption and the armor encoding come from the age
// standard library.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// version is the safebox release reported by the "version" command.
const version = "v0.2.0"

const armorHeader = "-----BEGIN AGE ENCRYPTED FILE-----"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, errUsage):
		// The command that rejected the command line already wrote a
		// diagnostic and the relevant usage text to stderr.
	case errors.Is(err, context.Canceled):
		fmt.Fprintln(os.Stderr, "safebox: "+tr(effectiveLanguage(), "diag.interrupted"))
	default:
		fmt.Fprintln(os.Stderr, "safebox: "+err.Error())
	}
	os.Exit(exitCode(err))
}

// exitCode maps a run error to the process exit status: 0 success, 2 malformed
// command line, 130 interrupted, 1 any other failure.
func exitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUsage):
		return 2
	case errors.Is(err, context.Canceled):
		return 130
	default:
		return 1
	}
}

// errUsage signals that the command line was malformed; main maps it to exit 2.
var errUsage = errors.New("usage error")

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	lang, args := extractLanguage(args)
	if len(args) == 0 {
		return printUsageError(stderr, func() { usage(stderr, lang) }, tr(lang, "diag.missingCommand"))
	}

	switch args[0] {
	case "keygen":
		return cmdKeygen(ctx, args[1:], stdout, stderr, lang)
	case "encrypt":
		return cmdEncrypt(ctx, args[1:], stdin, stdout, stderr, lang)
	case "decrypt":
		return cmdDecrypt(ctx, args[1:], stdin, stdout, stderr, lang)
	case "version", "--version", "-version":
		fmt.Fprintf(stdout, "safebox %s\n", version)
		return nil
	case "help", "-h", "-help", "--help":
		usage(stdout, lang)
		return nil
	default:
		return printUsageError(stderr, func() { usage(stderr, lang) }, tr(lang, "diag.unknownCommand", args[0]))
	}
}

// extractLanguage resolves the UI language and strips an optional global
// --lang flag from the front of args. The flag is only recognized before the
// command, so subcommand flags and their semantics stay exactly as before.
func extractLanguage(args []string) (string, []string) {
	lang := effectiveLanguage()
	for len(args) > 0 {
		a := args[0]
		switch {
		case a == "--lang" || a == "-lang":
			if len(args) < 2 {
				return lang, args
			}
			lang = normalizeLanguage(args[1])
			args = args[2:]
		case strings.HasPrefix(a, "--lang="):
			lang = normalizeLanguage(strings.TrimPrefix(a, "--lang="))
			args = args[1:]
		case strings.HasPrefix(a, "-lang="):
			lang = normalizeLanguage(strings.TrimPrefix(a, "-lang="))
			args = args[1:]
		default:
			return lang, args
		}
	}
	return lang, args
}

// printUsageError writes a single actionable diagnostic line followed by the
// relevant usage text to w, then returns errUsage so main exits with status 2.
// The diagnostic comes first so it stays visible when usage is long. message is
// already localized and formatted by the caller.
func printUsageError(w io.Writer, usageFn func(), message string) error {
	fmt.Fprintln(w, "safebox: "+message)
	if usageFn != nil {
		usageFn()
	}
	return errUsage
}

func usage(w io.Writer, lang string) {
	fmt.Fprintf(w, "%s\n\n%s\n%s\n\n%s\n%s\n%s\n%s\n%s\n%s\n\n%s\n\n%s\n%s\n",
		tr(lang, "usage.title", version),
		tr(lang, "usage.header"),
		tr(lang, "usage.synopsis"),
		tr(lang, "usage.commands"),
		tr(lang, "usage.keygen"),
		tr(lang, "usage.encrypt"),
		tr(lang, "usage.decrypt"),
		tr(lang, "usage.version"),
		tr(lang, "usage.help"),
		tr(lang, "usage.commandHelp"),
		tr(lang, "usage.plaintext"),
		tr(lang, "usage.lang"),
	)
}

type stringsFlag []string

func (s *stringsFlag) String() string { return strings.Join(*s, ",") }
func (s *stringsFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// parseCommandFlags parses a subcommand's flags. It reports help=true when the
// user asked for -h/--help (the flag package has already printed the usage),
// and maps every other parse failure to a usage error (exit status 2).
func parseCommandFlags(fs *flag.FlagSet, args []string) (help bool, err error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, nil
		}
		// The flag package already wrote the specific parse error and the
		// usage text to the configured output (stderr).
		return false, errUsage
	}
	return false, nil
}

func cmdKeygen(_ context.Context, args []string, stdout, stderr io.Writer, lang string) error {
	fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
	fs.SetOutput(stderr)
	identity := fs.String("identity", "", tr(lang, "flag.keygen.identity"))
	fs.Usage = func() {
		fmt.Fprintln(stderr, tr(lang, "keygen.usageLine"))
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, tr(lang, "keygen.detail"))
		fs.PrintDefaults()
	}
	help, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if help {
		return nil
	}
	if fs.NArg() > 0 {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.unexpectedArgument", "keygen", fs.Arg(0)))
	}
	if *identity == "" {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.keygen.requiresIdentity"))
	}
	if err := os.MkdirAll(filepath.Dir(*identity), 0o700); err != nil {
		return fmt.Errorf("creating key directory: %w", err)
	}

	pubPath := *identity + ".pub"
	for _, p := range []string{*identity, pubPath} {
		if err := ensureAbsent(p); err != nil {
			return err
		}
	}

	id, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("generating identity: %w", err)
	}

	if err := writeNewFile(*identity, 0o600, []byte(id.String()+"\n")); err != nil {
		return err
	}
	if err := writeNewFile(pubPath, 0o644, []byte(id.Recipient().String()+"\n")); err != nil {
		os.Remove(*identity)
		return err
	}

	fmt.Fprintf(stderr, "identity written: %s\n", *identity)
	fmt.Fprintf(stderr, "recipient written: %s\n", pubPath)
	fmt.Fprintln(stdout, id.Recipient().String())
	return nil
}

func cmdEncrypt(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, lang string) error {
	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var recipients stringsFlag
	fs.Var(&recipients, "recipient", tr(lang, "flag.encrypt.recipient"))
	in := fs.String("in", "", tr(lang, "flag.encrypt.in"))
	out := fs.String("out", "", tr(lang, "flag.encrypt.out"))
	fs.Usage = func() {
		fmt.Fprintln(stderr, tr(lang, "encrypt.usageLine"))
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, tr(lang, "encrypt.detail"))
		fs.PrintDefaults()
	}
	help, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if help {
		return nil
	}
	if fs.NArg() > 0 {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.unexpectedArgument", "encrypt", fs.Arg(0)))
	}
	if len(recipients) == 0 {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.encrypt.requiresRecipient"))
	}

	recips := make([]age.Recipient, 0, len(recipients))
	for _, r := range recipients {
		rr, err := age.ParseX25519Recipient(strings.TrimSpace(r))
		if err != nil {
			return fmt.Errorf("invalid recipient %q: %w", r, err)
		}
		recips = append(recips, rr)
	}

	inR, closeIn, err := openInput(*in, stdin)
	if err != nil {
		return err
	}
	defer closeIn()

	outW, commit, err := openOutput(*out, stdout)
	if err != nil {
		return err
	}

	armorW := armor.NewWriter(outW)
	enc, err := age.Encrypt(armorW, recips...)
	if err != nil {
		commit(false)
		return fmt.Errorf("encrypting: %w", err)
	}
	if _, err := io.Copy(enc, &ctxReader{ctx: ctx, r: inR}); err != nil {
		enc.Close()
		commit(false)
		return wrapCancel(err)
	}
	if err := enc.Close(); err != nil {
		commit(false)
		return fmt.Errorf("encrypting: %w", err)
	}
	if err := armorW.Close(); err != nil {
		commit(false)
		return fmt.Errorf("finishing armor: %w", err)
	}
	if err := commit(true); err != nil {
		return fmt.Errorf("committing encrypted output: %w", err)
	}
	return nil
}

func cmdDecrypt(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, lang string) error {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	identity := fs.String("identity", "", tr(lang, "flag.decrypt.identity"))
	in := fs.String("in", "", tr(lang, "flag.decrypt.in"))
	out := fs.String("out", "", tr(lang, "flag.decrypt.out"))
	fs.Usage = func() {
		fmt.Fprintln(stderr, tr(lang, "decrypt.usageLine"))
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, tr(lang, "decrypt.detail"))
		fs.PrintDefaults()
	}
	help, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if help {
		return nil
	}
	if fs.NArg() > 0 {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.unexpectedArgument", "decrypt", fs.Arg(0)))
	}
	if *identity == "" {
		return printUsageError(stderr, fs.Usage, tr(lang, "diag.decrypt.requiresIdentity"))
	}
	// Reject an existing --out target before doing any decryption or creating
	// the temporary plaintext file. openOutput's O_EXCL remains the
	// authoritative guard against a race between this check and creation.
	if err := checkOutputAbsent(*out); err != nil {
		return err
	}

	keyData, err := os.ReadFile(*identity)
	if err != nil {
		return fmt.Errorf("reading identity: %w", err)
	}
	// ParseIdentities accepts age-keygen style files: one or more identities,
	// with blank lines and "#" comments (e.g. "# public key: age1...") ignored.
	ids, err := age.ParseIdentities(bytes.NewReader(keyData))
	if err != nil {
		return fmt.Errorf("invalid identity file %s: %w", *identity, err)
	}

	inR, closeIn, err := openInput(*in, stdin)
	if err != nil {
		return err
	}
	defer closeIn()

	br := bufio.NewReader(inR)
	var src io.Reader = br
	if isArmored(br) {
		src = armor.NewReader(br)
	}

	dec, err := age.Decrypt(src, ids...)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}
	// Age authenticates the complete stream only at EOF. Spool to a private
	// temporary file so large inputs don't require an unbounded in-memory buffer,
	// and don't expose any plaintext until the full stream has been verified.
	tmp, err := os.CreateTemp("", "safebox-plaintext-*")
	if err != nil {
		return fmt.Errorf("creating temporary plaintext file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()
	if _, err := io.Copy(tmp, &ctxReader{ctx: ctx, r: dec}); err != nil {
		return wrapCancel(fmt.Errorf("decryption failed: %w", err))
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("syncing temporary plaintext: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewinding temporary plaintext: %w", err)
	}

	outW, commit, err := openOutput(*out, stdout)
	if err != nil {
		return err
	}
	if _, err := io.Copy(outW, tmp); err != nil {
		commit(false)
		return fmt.Errorf("writing output: %w", err)
	}
	if err := commit(true); err != nil {
		return fmt.Errorf("committing decrypted output: %w", err)
	}
	return nil
}

func wrapCancel(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return context.Canceled
	}
	return err
}

// openInput returns a reader for path, or def when path is empty or "-".
func openInput(path string, def io.Reader) (io.Reader, func(), error) {
	if path == "" || path == "-" {
		return def, func() {}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening input: %w", err)
	}
	return f, func() { f.Close() }, nil
}

// openOutput returns a writer for path, or def when path is empty or "-".
//
// Files are created exclusively (O_EXCL) with mode 0600 and are removed again
// if commit(false) is called, so no partial or unintended output is left behind.
func openOutput(path string, def io.Writer) (io.Writer, func(success bool) error, error) {
	if path == "" || path == "-" {
		return def, func(bool) error { return nil }, nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, nil, fmt.Errorf("refusing to overwrite existing file: %s", path)
		}
		return nil, nil, fmt.Errorf("creating output: %w", err)
	}
	commit := func(success bool) error {
		closeErr := f.Close()
		if !success || closeErr != nil {
			removeErr := os.Remove(path)
			if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				if closeErr != nil {
					return errors.Join(fmt.Errorf("closing output: %w", closeErr), fmt.Errorf("removing partial output: %w", removeErr))
				}
				return fmt.Errorf("removing partial output: %w", removeErr)
			}
		}
		if closeErr != nil {
			return fmt.Errorf("closing output: %w", closeErr)
		}
		return nil
	}
	return f, commit, nil
}

func ensureAbsent(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("refusing to overwrite existing file: %s", path)
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// checkOutputAbsent refuses when path already exists, using Lstat so symlinks
// are not followed and a dangling symlink still counts as an existing entry.
// It is a best-effort early check; openOutput's O_EXCL is the authoritative
// guard against a race between this check and creation.
func checkOutputAbsent(path string) error {
	if path == "" || path == "-" {
		return nil
	}
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return fmt.Errorf("refusing to overwrite existing file: %s", path)
	case errors.Is(err, os.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("checking output path: %w", err)
	}
}

func writeNewFile(path string, perm os.FileMode, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("refusing to overwrite existing file: %s", path)
		}
		return fmt.Errorf("creating %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func isArmored(br *bufio.Reader) bool {
	b, err := br.Peek(len(armorHeader))
	if err != nil && len(b) < len(armorHeader) {
		return false
	}
	return string(b) == armorHeader
}

// ctxReader checks for cancellation between reads. An underlying Reader blocked
// in Read may not be interruptible until that read returns.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
	}
	return c.r.Read(p)
}
