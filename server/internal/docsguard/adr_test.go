package docsguard_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// adrDir is docs/decisions relative to this package.
const adrDir = "../../../docs/decisions"

// adrFile matches the required filename shape: a four-digit number, a
// hyphen, a lowercase-hyphenated slug, and .md.
var adrFile = regexp.MustCompile(`^(\d{4})-[a-z0-9]+(?:-[a-z0-9]+)*\.md$`)

// adrHeading matches the number in an ADR's H1. Both spellings in the
// tree are accepted — "# ADR 0044 — ..." and "# 0047 — ..." — because
// the number is the load-bearing part and reformatting forty files to
// settle a cosmetic difference is not this test's job.
var adrHeading = regexp.MustCompile(`^#\s+(?:ADR\s+)?(\d{4})\b`)

// TestADRNumbersAreUniqueAndMatchTheirHeading is the guard AGENTS.md
// §4 asks authors to enforce by hand.
//
// Two rules, and the first is the one that has actually been broken:
//
//  1. One file per number. Two branches picking "the next free
//     number" against a main that has neither produce different
//     FILENAMES, so git merges both without a conflict and neither
//     author learns. Three collisions reached main this way before
//     this test existed; the fix is that CI lists the directory,
//     which is the only place the problem is visible.
//
//  2. The number in the H1 matches the number in the filename. A
//     renumbered ADR whose heading still says the old number is worse
//     than a duplicate, because every inbound link looks right and
//     lands on a document that disagrees with itself.
func TestADRNumbersAreUniqueAndMatchTheirHeading(t *testing.T) {
	entries, err := os.ReadDir(adrDir)
	if err != nil {
		t.Fatalf("read %s: %v", adrDir, err)
	}

	byNumber := map[string][]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		// README and other prose in the directory are not ADRs.
		m := adrFile.FindStringSubmatch(e.Name())
		if m == nil {
			if strings.EqualFold(e.Name(), "README.md") {
				continue
			}
			t.Errorf("%s: filename is not NNNN-lowercase-slug.md — an ADR that does not sort with the others is an ADR nobody finds", e.Name())
			continue
		}
		num := m[1]
		byNumber[num] = append(byNumber[num], e.Name())

		raw, err := os.ReadFile(filepath.Join(adrDir, e.Name()))
		if err != nil {
			t.Errorf("%s: %v", e.Name(), err)
			continue
		}
		heading := firstHeading(string(raw))
		if heading == "" {
			t.Errorf("%s: no H1 of the form '# ADR %s — title' or '# %s — title'", e.Name(), num, num)
			continue
		}
		if heading != num {
			t.Errorf("%s: filename says ADR %s, H1 says ADR %s — AGENTS.md §4 requires both", e.Name(), num, heading)
		}
	}

	if len(byNumber) == 0 {
		t.Fatalf("no ADRs found under %s — the guard is pointed at the wrong directory", adrDir)
	}

	for num, files := range byNumber {
		if len(files) > 1 {
			t.Errorf("ADR %s is claimed by %d files: %s\n"+
				"Two branches each took the next free number against a main that had neither. "+
				"Renumber the one with fewer inbound links (grep for its filename), fix its H1, "+
				"and update AGENTS.md's ADR range line.", num, len(files), strings.Join(files, ", "))
		}
	}
}

// firstHeading returns the ADR number from the file's first H1, or ""
// if the first heading is not one.
func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		if m := adrHeading.FindStringSubmatch(line); m != nil {
			return m[1]
		}
		return ""
	}
	return ""
}
