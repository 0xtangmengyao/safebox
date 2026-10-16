package main

import (
	"regexp"
	"sort"
	"testing"
)

// formatVerb matches a printf-style verb such as %s, %q, %d or %v.
var formatVerb = regexp.MustCompile(`%[a-zA-Z]`)

func placeholderSet(s string) map[string]bool {
	set := map[string]bool{}
	for _, m := range formatVerb.FindAllString(s, -1) {
		set[m] = true
	}
	return set
}

// TestCatalogsCoverAllLanguages guards against a translation missing a key or
// a printf verb, which would otherwise ship as an empty or broken message.
func TestCatalogsCoverAllLanguages(t *testing.T) {
	en, ok := catalogs[defaultLanguage]
	if !ok {
		t.Fatalf("missing default catalog %q", defaultLanguage)
	}
	known := make([]string, 0, len(supportedLanguages))
	known = append(known, supportedLanguages...)
	for _, lang := range supportedLanguages {
		c, ok := catalogs[lang]
		if !ok {
			t.Fatalf("supported language %q has no catalog", lang)
		}
		for key, enMsg := range en {
			msg, ok := c[key]
			if !ok {
				t.Errorf("%s: missing key %q", lang, key)
				continue
			}
			if msg == "" {
				t.Errorf("%s: key %q is empty", lang, key)
			}
			want := placeholderSet(enMsg)
			got := placeholderSet(msg)
			if len(want) != len(got) {
				t.Errorf("%s: key %q placeholder mismatch: en=%v %s=%v", lang, key, sorted(want), lang, sorted(got))
			}
		}
		for key := range c {
			if _, ok := en[key]; !ok {
				t.Errorf("%s: unexpected key %q", lang, key)
			}
		}
	}
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestNormalizeLanguage(t *testing.T) {
	cases := map[string]string{
		"":            "en",
		"en":          "en",
		"EN":          "en",
		"en-US":       "en",
		"zh":          "zh-CN",
		"zh_CN":       "zh-CN",
		"zh_CN.UTF-8": "zh-CN",
		"zh-Hans":     "zh-CN",
		"ja":          "ja",
		"ja-JP":       "ja",
		"ko":          "ko",
		"ko_KR.UTF-8": "ko",
		"es":          "es",
		"es_MX":       "es",
		"fr":          "fr",
		"fr-FR":       "fr",
		"de":          "en",
		"  FR  ":      "fr",
	}
	for in, want := range cases {
		if got := normalizeLanguage(in); got != want {
			t.Errorf("normalizeLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTRFallsBackToEnglish(t *testing.T) {
	if got := tr("de", "usage.header"); got != enMessages["usage.header"] {
		t.Fatalf("unknown language fallback = %q", got)
	}
	if got := tr("fr", "no.such.key"); got != "no.such.key" {
		t.Fatalf("missing key = %q, want key itself", got)
	}
	if got := tr("en", "usage.title", "v9"); got != "safebox v9 - age-based file and text encryption" {
		t.Fatalf("formatted message = %q", got)
	}
}
