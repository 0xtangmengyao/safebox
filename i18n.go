package main

import (
	"fmt"
	"os"
	"strings"
)

// catalog maps message IDs to localized text. Message text may contain
// fmt-style verbs; callers pass arguments through tr. Keeping every catalog
// keyed by the same IDs lets i18n_test.go assert that no translation is
// missing, and lets tr fall back to English for unknown languages.
type catalog map[string]string

// defaultLanguage is used whenever no supported language is requested.
const defaultLanguage = "en"

// supportedLanguages are the shipped UI languages, in a stable order.
var supportedLanguages = []string{"en", "zh-CN", "ja", "ko", "es", "fr"}

// catalogs holds one catalog per supported language. English is authoritative:
// any key missing from another catalog falls back to English.
var catalogs = map[string]catalog{
	"en":    enMessages,
	"zh-CN": zhMessages,
	"ja":    jaMessages,
	"ko":    koMessages,
	"es":    esMessages,
	"fr":    frMessages,
}

// normalizeLanguage maps locale-like input (for example "zh_CN.UTF-8",
// "ja-JP" or "FR") to one of the supported language codes. Unknown values
// resolve to English.
func normalizeLanguage(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultLanguage
	}
	// Strip an encoding suffix such as ".UTF-8" and normalize separators.
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	s = strings.ToLower(strings.ReplaceAll(s, "_", "-"))
	switch {
	case strings.HasPrefix(s, "zh"):
		return "zh-CN"
	case strings.HasPrefix(s, "ja"):
		return "ja"
	case strings.HasPrefix(s, "ko"):
		return "ko"
	case strings.HasPrefix(s, "es"):
		return "es"
	case strings.HasPrefix(s, "fr"):
		return "fr"
	case strings.HasPrefix(s, "en"):
		return "en"
	default:
		return defaultLanguage
	}
}

// envLanguage returns the language requested via SAFEBOX_LANG, or "" when the
// variable is unset or holds an unsupported value.
func envLanguage() string {
	v := strings.TrimSpace(os.Getenv("SAFEBOX_LANG"))
	if v == "" {
		return ""
	}
	for _, code := range supportedLanguages {
		if strings.EqualFold(v, code) {
			return code
		}
	}
	// Accept locale-like aliases too, but only if they are meaningful.
	norm := normalizeLanguage(v)
	if norm != defaultLanguage || strings.HasPrefix(strings.ToLower(v), "en") {
		return norm
	}
	return ""
}

// effectiveLanguage resolves the language used when no --lang flag is given:
// SAFEBOX_LANG when set to a supported value, otherwise English.
func effectiveLanguage() string {
	if lang := envLanguage(); lang != "" {
		return lang
	}
	return defaultLanguage
}

// tr looks up key in lang, falling back to English and finally to the key
// itself. When args are supplied the message is treated as a format string.
func tr(lang, key string, args ...any) string {
	msg, ok := lookup(lang, key)
	if !ok {
		msg = key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func lookup(lang, key string) (string, bool) {
	if c, ok := catalogs[lang]; ok {
		if msg, ok := c[key]; ok {
			return msg, true
		}
	}
	if msg, ok := catalogs[defaultLanguage][key]; ok {
		return msg, true
	}
	return "", false
}
