package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"filippo.io/age"
)

// TestMain lets the test binary re-execute itself as the safebox CLI so tests
// can assert the real process exit status, not just the in-process error.
func TestMain(m *testing.M) {
	if os.Getenv("SAFEBOX_RUN_MAIN") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runCLI(t *testing.T, stdin []byte, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err = run(context.Background(), args, bytes.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), err
}

// keygenFiles generates a key pair in a fresh temp dir and returns the paths
// plus the recipient string.
func keygenFiles(t *testing.T) (idPath, pubPath, recipient string) {
	t.Helper()
	dir := t.TempDir()
	idPath = filepath.Join(dir, "identity.txt")
	out, _, err := runCLI(t, nil, "keygen", "--identity", idPath)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return idPath, idPath + ".pub", strings.TrimSpace(out)
}

func TestKeygenWritesValidKeys(t *testing.T) {
	idPath, pubPath, recipient := keygenFiles(t)

	idData, err := os.ReadFile(idPath)
	if err != nil {
		t.Fatalf("read identity: %v", err)
	}
	id, err := age.ParseX25519Identity(strings.TrimSpace(string(idData)))
	if err != nil {
		t.Fatalf("identity does not parse: %v", err)
	}
	pubData, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatalf("read recipient: %v", err)
	}
	pub := strings.TrimSpace(string(pubData))
	if pub != id.Recipient().String() {
		t.Fatalf("recipient %q != identity recipient %q", pub, id.Recipient().String())
	}
	if pub != recipient {
		t.Fatalf("stdout recipient %q != file recipient %q", recipient, pub)
	}

	if runtime.GOOS != "windows" {
		assertPerm(t, idPath, 0o600)
		// umask may strip bits from the requested mode, so the public key may
		// be narrower than 0644 but must never be wider.
		assertPermAtMost(t, pubPath, 0o644)
	}
}

func TestKeygenCreatesPrivateParentDirectory(t *testing.T) {
	dir := t.TempDir()
	idPath := filepath.Join(dir, "keys", "alice.txt")
	if _, _, err := runCLI(t, nil, "keygen", "--identity", idPath); err != nil {
		t.Fatalf("keygen in new directory: %v", err)
	}
	if _, err := os.Stat(idPath); err != nil {
		t.Fatalf("identity file not created: %v", err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Dir(idPath))
		if err != nil {
			t.Fatalf("stat key directory: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Fatalf("key directory mode = %o, want 700", got)
		}
	}
}

func TestKeygenRefusesOverwrite(t *testing.T) {
	idPath, pubPath, _ := keygenFiles(t)

	if _, _, err := runCLI(t, nil, "keygen", "--identity", idPath); err == nil {
		t.Fatal("keygen overwrote existing identity; want error")
	}
	// Simulate a stray public key without a private key.
	if err := os.Remove(idPath); err != nil {
		t.Fatalf("remove identity: %v", err)
	}
	if _, _, err := runCLI(t, nil, "keygen", "--identity", idPath); err == nil {
		t.Fatal("keygen overwrote existing .pub; want error")
	}
	// The failed run must not have recreated the private key.
	if _, err := os.Stat(idPath); !os.IsNotExist(err) {
		t.Fatalf("identity file unexpectedly present after failed keygen: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Fatalf("existing .pub removed: %v", err)
	}
}

func TestRoundTripStdinStdout(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	plaintext := []byte("hello safebox\nsecond line\n")

	ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !strings.HasPrefix(ciphertext, armorHeader) {
		t.Fatalf("output is not armor text:\n%s", ciphertext)
	}
	if strings.Contains(ciphertext, string(plaintext)) {
		t.Fatal("ciphertext contains plaintext")
	}

	got, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", idPath)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != string(plaintext) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plaintext)
	}
}

func TestRoundTripFiles(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	dir := t.TempDir()
	inPath := filepath.Join(dir, "plain.txt")
	encPath := filepath.Join(dir, "secret.age")
	decPath := filepath.Join(dir, "out.txt")
	plaintext := []byte("file round trip \x00 binary ok")

	if err := os.WriteFile(inPath, plaintext, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if _, _, err := runCLI(t, nil, "encrypt", "--recipient", recipient, "--in", inPath, "--out", encPath); err != nil {
		t.Fatalf("encrypt to file: %v", err)
	}
	if _, _, err := runCLI(t, nil, "decrypt", "--identity", idPath, "--in", encPath, "--out", decPath); err != nil {
		t.Fatalf("decrypt to file: %v", err)
	}

	got, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plaintext)
	}

	if runtime.GOOS != "windows" {
		assertPerm(t, encPath, 0o600)
		assertPerm(t, decPath, 0o600)
	}
}

func TestDecryptLargePayloadAndCleansTemporaryPlaintext(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	plaintext := bytes.Repeat([]byte("large encrypted payload\n"), 512*1024)
	ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)
	got, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", idPath)
	if err != nil {
		t.Fatalf("decrypt large payload: %v", err)
	}
	if !bytes.Equal([]byte(got), plaintext) {
		t.Fatal("large payload round-trip mismatch")
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read temporary directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary plaintext was not removed: %v", entries)
	}
}

func TestMultipleRecipients(t *testing.T) {
	id1, _, rec1 := keygenFiles(t)
	id2, _, rec2 := keygenFiles(t)
	plaintext := []byte("shared message")

	ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", rec1, "--recipient", rec2)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	for _, id := range []string{id1, id2} {
		got, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", id)
		if err != nil {
			t.Fatalf("decrypt with %s: %v", id, err)
		}
		if got != string(plaintext) {
			t.Fatalf("mismatch with %s: got %q", id, got)
		}
	}
}

func TestTamperedCiphertextFailsWithoutPlaintext(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	plaintext := []byte("sensitive payload")

	ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	lines := strings.Split(ciphertext, "\n")
	tampered := false
	for i, ln := range lines {
		if len(ln) > 10 && !strings.HasPrefix(ln, "-") {
			b := []byte(ln)
			if b[0] == 'A' {
				b[0] = 'B'
			} else {
				b[0] = 'A'
			}
			lines[i] = string(b)
			tampered = true
			break
		}
	}
	if !tampered {
		t.Fatal("could not find a body line to tamper with")
	}

	got, _, err := runCLI(t, []byte(strings.Join(lines, "\n")), "decrypt", "--identity", idPath)
	if err == nil {
		t.Fatal("tampered ciphertext decrypted; want error")
	}
	if strings.Contains(got, string(plaintext)) {
		t.Fatalf("plaintext leaked on failure: %q", got)
	}
	if got != "" {
		t.Fatalf("output written on failure: %q", got)
	}
}

func TestWrongIdentityFailsWithoutPlaintext(t *testing.T) {
	_, _, recipient := keygenFiles(t)
	wrongID, _, _ := keygenFiles(t)
	plaintext := []byte("only for the real owner")

	ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	got, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", wrongID)
	if err == nil {
		t.Fatal("wrong identity decrypted; want error")
	}
	if got != "" {
		t.Fatalf("output written on failure: %q", got)
	}
}

func TestDecryptBinaryAgeInput(t *testing.T) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	dir := t.TempDir()
	idPath := filepath.Join(dir, "id.txt")
	if err := os.WriteFile(idPath, []byte(id.String()+"\n"), 0o600); err != nil {
		t.Fatalf("write identity: %v", err)
	}

	plaintext := []byte("binary age input")
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, id.Recipient())
	if err != nil {
		t.Fatalf("age.Encrypt: %v", err)
	}
	if _, err := w.Write(plaintext); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	got, _, err := runCLI(t, buf.Bytes(), "decrypt", "--identity", idPath)
	if err != nil {
		t.Fatalf("decrypt binary: %v", err)
	}
	if got != string(plaintext) {
		t.Fatalf("mismatch: got %q", got)
	}
}

func TestNoOverwriteTargets(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(existing, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	if _, _, err := runCLI(t, []byte("data"), "encrypt", "--recipient", recipient, "--out", existing); err == nil {
		t.Fatal("encrypt overwrote existing file; want error")
	}
	ciphertext, _, err := runCLI(t, []byte("data"), "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", idPath, "--out", existing); err == nil {
		t.Fatal("decrypt overwrote existing file; want error")
	}

	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("read existing: %v", err)
	}
	if string(got) != "keep me" {
		t.Fatalf("existing file modified: %q", got)
	}
}

func TestInvalidRecipient(t *testing.T) {
	if _, _, err := runCLI(t, []byte("x"), "encrypt", "--recipient", "not-a-key"); err == nil {
		t.Fatal("invalid recipient accepted; want error")
	}
}

func TestUsageAndVersion(t *testing.T) {
	out, _, err := runCLI(t, nil, "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, version) {
		t.Fatalf("version output %q missing %q", out, version)
	}

	out, _, err = runCLI(t, nil, "help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out, "safebox") {
		t.Fatalf("help output missing usage: %q", out)
	}

	if _, _, err := runCLI(t, nil, "bogus"); err == nil {
		t.Fatal("unknown command accepted; want error")
	}
}

func TestDecryptIdentityFileCommentsAndMultipleIdentities(t *testing.T) {
	id1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity 1: %v", err)
	}
	id2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity 2: %v", err)
	}

	dir := t.TempDir()
	idPath := filepath.Join(dir, "keys.txt")
	content := fmt.Sprintf(
		"# created: 2026-01-01T00:00:00Z\n"+
			"# public key: %s\n"+
			"%s\n"+
			"\n"+
			"# second identity\n"+
			"%s\n",
		id1.Recipient(), id1, id2)
	if err := os.WriteFile(idPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write identity file: %v", err)
	}

	plaintext := []byte("multi identity message")
	for _, id := range []*age.X25519Identity{id1, id2} {
		ciphertext, _, err := runCLI(t, plaintext, "encrypt", "--recipient", id.Recipient().String())
		if err != nil {
			t.Fatalf("encrypt to %s: %v", id.Recipient(), err)
		}
		got, _, err := runCLI(t, []byte(ciphertext), "decrypt", "--identity", idPath)
		if err != nil {
			t.Fatalf("decrypt with %s: %v", id.Recipient(), err)
		}
		if got != string(plaintext) {
			t.Fatalf("round-trip mismatch with %s: got %q", id.Recipient(), got)
		}
	}
}

func TestSubcommandHelpSucceeds(t *testing.T) {
	for _, cmd := range []string{"keygen", "encrypt", "decrypt"} {
		for _, h := range []string{"-h", "--help"} {
			stdout, stderr, err := runCLI(t, nil, cmd, h)
			if err != nil {
				t.Fatalf("%s %s: want success, got %v", cmd, h, err)
			}
			if stdout != "" {
				t.Fatalf("%s %s: unexpected stdout %q", cmd, h, stdout)
			}
			if !strings.Contains(stderr, "Usage:") {
				t.Fatalf("%s %s: usage not printed:\n%s", cmd, h, stderr)
			}
		}
	}
}

func TestSubcommandBadFlagsAreUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"keygen unknown flag", []string{"keygen", "--bogus"}},
		{"encrypt unknown flag", []string{"encrypt", "--bogus"}},
		{"decrypt unknown flag", []string{"decrypt", "--bogus"}},
		{"keygen extra arg", []string{"keygen", "extra"}},
		{"encrypt extra arg", []string{"encrypt", "extra"}},
		{"decrypt extra arg", []string{"decrypt", "extra"}},
	}
	for _, tc := range cases {
		_, _, err := runCLI(t, nil, tc.args...)
		if err == nil {
			t.Fatalf("%s: want error", tc.name)
		}
		if !errors.Is(err, errUsage) {
			t.Fatalf("%s: want usage error, got %v", tc.name, err)
		}
	}
}

// usageErrorCases are malformed command lines that must all be rejected with a
// diagnostic, usage text and exit status 2. Previously the missing-required-flag
// and extra-argument cases exited 2 with empty stderr. diag is the diagnostic
// substring each case must print to stderr.
var usageErrorCases = []struct {
	name string
	args []string
	diag string
}{
	{"no command", nil, "safebox: missing command"},
	{"keygen missing required flag", []string{"keygen"}, "safebox: keygen requires --identity PATH"},
	{"encrypt missing required flag", []string{"encrypt"}, "safebox: encrypt requires at least one --recipient"},
	{"decrypt missing required flag", []string{"decrypt"}, "safebox: decrypt requires --identity PATH"},
	{"keygen extra arg", []string{"keygen", "extra"}, `safebox: keygen: unexpected argument "extra"`},
	{"encrypt extra arg", []string{"encrypt", "extra"}, `safebox: encrypt: unexpected argument "extra"`},
	{"decrypt extra arg", []string{"decrypt", "extra"}, `safebox: decrypt: unexpected argument "extra"`},
	{"unknown command", []string{"bogus"}, `safebox: unknown command "bogus"`},
	{"keygen unknown flag", []string{"keygen", "--bogus"}, "flag provided but not defined: -bogus"},
	{"encrypt unknown flag", []string{"encrypt", "--bogus"}, "flag provided but not defined: -bogus"},
	{"decrypt unknown flag", []string{"decrypt", "--bogus"}, "flag provided but not defined: -bogus"},
}

func TestUsageErrorsPrintDiagnosticsAndUsage(t *testing.T) {
	for _, tc := range usageErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, nil, tc.args...)
			if err == nil {
				t.Fatal("want error")
			}
			if !errors.Is(err, errUsage) {
				t.Fatalf("want usage error, got %v", err)
			}
			if code := exitCode(err); code != 2 {
				t.Fatalf("exitCode = %d, want 2", code)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(stderr, tc.diag) {
				t.Fatalf("stderr %q lacks diagnostic %q", stderr, tc.diag)
			}
			if !strings.Contains(stderr, "Usage:") {
				t.Fatalf("stderr lacks usage text: %q", stderr)
			}
		})
	}
}

func TestExitCodeMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"success", nil, 0},
		{"usage", errUsage, 2},
		{"wrapped usage", fmt.Errorf("wrap: %w", errUsage), 2},
		{"interrupted", context.Canceled, 130},
		{"wrapped interrupted", fmt.Errorf("wrap: %w", context.Canceled), 130},
		{"failure", errors.New("boom"), 1},
	}
	for _, tc := range cases {
		if got := exitCode(tc.err); got != tc.want {
			t.Fatalf("%s: exitCode = %d, want %d", tc.name, got, tc.want)
		}
	}
}

// TestUsageErrorsExitTwoWithoutStdout re-executes the test binary as the CLI to
// verify the real process exit status, stdout and stderr contract.
func TestUsageErrorsExitTwoWithoutStdout(t *testing.T) {
	if os.Getenv("SAFEBOX_RUN_MAIN") == "1" {
		t.Skip("helper process")
	}
	for _, tc := range usageErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], tc.args...)
			cmd.Env = append(os.Environ(), "SAFEBOX_RUN_MAIN=1")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("want non-zero exit, got err=%v", err)
			}
			if code := exitErr.ExitCode(); code != 2 {
				t.Fatalf("exit code = %d, want 2", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), tc.diag) || !strings.Contains(stderr.String(), "Usage:") {
				t.Fatalf("stderr %q lacks diagnostic %q or usage", stderr.String(), tc.diag)
			}
		})
	}
}

func TestDecryptRejectsExistingOutputBeforeDecrypting(t *testing.T) {
	idPath, _, _ := keygenFiles(t)
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(existing, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	// The input is not a valid age file: if the output-exists check runs before
	// decryption we see an overwrite error, not a decryption error.
	_, _, err := runCLI(t, []byte("this is not an age file"), "decrypt", "--identity", idPath, "--out", existing)
	if err == nil {
		t.Fatal("decrypt wrote to existing file; want error")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("want overwrite rejection before decryption, got: %v", err)
	}
	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("read existing: %v", err)
	}
	if string(got) != "keep me" {
		t.Fatalf("existing file modified: %q", got)
	}
}

func TestDecryptRejectsDanglingSymlinkOutput(t *testing.T) {
	idPath, _, recipient := keygenFiles(t)
	ciphertext, _, err := runCLI(t, []byte("data"), "encrypt", "--recipient", recipient)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	dir := t.TempDir()
	link := filepath.Join(dir, "out-link")
	target := filepath.Join(dir, "target-does-not-exist")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, _, err = runCLI(t, []byte(ciphertext), "decrypt", "--identity", idPath, "--out", link)
	if err == nil {
		t.Fatal("decrypt through dangling symlink; want error")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("want overwrite rejection, got: %v", err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("symlink was removed: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink replaced by a regular file: %v", info.Mode())
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("symlink target unexpectedly created: %v", err)
	}
}

func assertPerm(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %o, want %o", path, got, want)
	}
}

// assertPermAtMost fails if path grants any permission bits outside max. umask
// may strip bits from the requested mode, so the actual mode may be narrower
// than max but must never be wider.
func assertPermAtMost(t *testing.T, path string, max os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	got := info.Mode().Perm()
	if got&^max != 0 {
		t.Fatalf("%s mode = %o, want no wider than %o", path, got, max)
	}
}
