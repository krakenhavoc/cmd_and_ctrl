package roadmap

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// closed.go — docs/engine-seams.md's Closed seams list, assembled from
// one fragment file per closure (#1461, Discussion #1231 Option A).
//
// The list used to be hand-written, and every engine PR added its entry
// at the top of it, so any two open PRs conflicted on the same lines
// even when their code never met. Now a PR that closes a seam adds
// docs/engine-seams/closed/<issue>-<slug>.md and nothing else; two PRs
// never write to the same file. The block between the CLOSED SEAMS
// markers is generated from the fragments by
//
//	go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed
//
// which CI runs on every push to develop and main (the census-publish
// job in .github/workflows/ci-cd.yml), exactly as it refreshes the
// catalog census. A PR does not run it: a regenerated block in a PR is
// the one shared hunk this file exists to remove. That is also why the
// flag is -update-closed and not the package's -update, which a PR
// that edits registry.go does run for the Open seams table.
//
// A fragment is a front-matter block and the entry's markdown:
//
//	---
//	title: "Phasing"
//	date: 2026-09-23
//	issues: [1199]
//	pr: 1251
//	---
//	**Phasing** (#1199, [ADR 0084](decisions/0084-phasing.md)) — …
//
// The front matter is a strict subset of YAML, so GitHub renders it as a
// table, but it is parsed here line by line and nothing else is
// accepted: title is a double-quoted (JSON) string, date is YYYY-MM-DD,
// issues is a JSON array of issue numbers (empty when the entry names
// none), pr is optional. The body must begin with the title in bold;
// the generator adds the "- " list marker. Links are written relative
// to docs/, because the body is rendered in docs/engine-seams.md.

// Generated-block delimiters for the closed-seams list.
const (
	ClosedBeginMarker = "<!-- BEGIN GENERATED CLOSED SEAMS — generated from docs/engine-seams/closed/*.md and refreshed by CI on every push to develop and main; add a fragment there instead of editing this list (manual refresh: go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed) -->"
	ClosedEndMarker   = "<!-- END GENERATED CLOSED SEAMS -->"
)

// ClosedSeamsDir is the fragment directory relative to this package's
// directory (go test's working directory).
const ClosedSeamsDir = "../../../docs/engine-seams/closed"

// LegacyClosedSeams is how many entries the hand-written list held when
// it was split into fragments (#1461). Those fragments carry
// legacy_order, their 1-based position in that list, and render after
// every newer fragment in exactly that order: dozens of them say "the
// row above" or "the entry below", so sorting them by date would
// falsify their prose without changing a word of it. No other fragment
// may carry legacy_order; the bound is what enforces that.
const LegacyClosedSeams = 134

// ClosedSeam is one parsed fragment.
type ClosedSeam struct {
	// File is the fragment's base name, e.g. "1199-phasing.md".
	File string
	// Title is the entry's bold heading, without the asterisks.
	Title string
	// Date is the day the closure landed, YYYY-MM-DD.
	Date string
	// Issues are the issue numbers the entry closes, in the order its
	// heading names them. Empty when it names none.
	Issues []int
	// PR is the pull request that closed the seam, 0 when not recorded.
	PR int
	// LegacyOrder is the position in the pre-fragment hand-written
	// list, 1..LegacyClosedSeams, or 0 for a fragment written since.
	LegacyOrder int
	// Body is the entry's markdown, starting with "**Title**", with no
	// list marker and no trailing newline.
	Body string
}

var (
	fragmentName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*\.md$`)
	fragmentDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// ParseClosedSeam parses one fragment. name is the file's base name and
// is used in every error, so a malformed fragment is named when the
// build fails on it.
func ParseClosedSeam(name string, data []byte) (ClosedSeam, error) {
	fail := func(format string, args ...any) (ClosedSeam, error) {
		return ClosedSeam{}, fmt.Errorf("closed-seam fragment %s: %s", name, fmt.Sprintf(format, args...))
	}
	if !fragmentName.MatchString(name) {
		return fail("the file name must be lower-case letters, digits and hyphens ending in .md (<issue>-<slug>.md)")
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return fail("must start with a front-matter line \"---\"")
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	var header string
	switch {
	case end >= 0:
		header, rest = rest[:end], rest[end+len("\n---\n"):]
	case strings.HasSuffix(rest, "\n---"):
		header, rest = strings.TrimSuffix(rest, "\n---"), ""
	default:
		return fail("the front matter has no closing \"---\" line")
	}

	s := ClosedSeam{File: name}
	seen := map[string]bool{}
	for i, line := range strings.Split(header, "\n") {
		key, value, ok := strings.Cut(line, ":")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" {
			return fail("front-matter line %d is not \"key: value\": %q", i+2, line)
		}
		if seen[key] {
			return fail("front-matter key %q appears twice", key)
		}
		seen[key] = true
		switch key {
		case "title":
			if err := json.Unmarshal([]byte(value), &s.Title); err != nil || !strings.HasPrefix(value, `"`) {
				return fail(`title must be a double-quoted string, e.g. title: "Phasing"; got %s`, value)
			}
			if strings.TrimSpace(s.Title) != s.Title || s.Title == "" || strings.Contains(s.Title, "**") {
				return fail("title must be non-empty, without leading or trailing spaces, and must not contain \"**\": %q", s.Title)
			}
		case "date":
			if !fragmentDate.MatchString(value) {
				return fail("date must be YYYY-MM-DD, got %q", value)
			}
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return fail("date %q is not a calendar date", value)
			}
			s.Date = value
		case "issues":
			if err := json.Unmarshal([]byte(value), &s.Issues); err != nil || !strings.HasPrefix(value, "[") {
				return fail("issues must be a list of issue numbers, e.g. issues: [1199] or issues: [] ; got %s", value)
			}
			for _, n := range s.Issues {
				if n <= 0 {
					return fail("issues must be positive issue numbers, got %d", n)
				}
			}
		case "pr":
			n, err := strconv.Atoi(value)
			if err != nil || n <= 0 {
				return fail("pr must be a pull-request number, got %q", value)
			}
			s.PR = n
		case "legacy_order":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 || n > LegacyClosedSeams {
				return fail("legacy_order is only for the %d entries migrated from the hand-written list (1..%d), got %q; a new fragment must not set it",
					LegacyClosedSeams, LegacyClosedSeams, value)
			}
			s.LegacyOrder = n
		default:
			return fail("unknown front-matter key %q (known: title, date, issues, pr, legacy_order)", key)
		}
	}
	for _, key := range []string{"title", "date", "issues"} {
		if !seen[key] {
			return fail("front matter is missing %q", key)
		}
	}

	s.Body = strings.Trim(rest, "\n")
	if strings.TrimSpace(s.Body) == "" {
		return fail("has no entry text after the front matter")
	}
	if want := "**" + s.Title + "**"; !strings.HasPrefix(s.Body, want) {
		return fail("the entry must begin with its title in bold, %s, and without a list marker (the generator adds \"- \")", want)
	}
	if strings.Contains(s.Body, "GENERATED CLOSED SEAMS") {
		return fail("the entry text contains a generated-block marker")
	}
	return s, nil
}

// LoadClosedSeams parses every fragment in fsys's root and returns them
// in list order (see SortClosedSeams). Anything in the directory that is
// not a well-formed fragment is an error, as are two fragments with the
// same title or the same legacy_order: a list that silently dropped or
// merged an entry is the failure this whole file exists to prevent.
func LoadClosedSeams(fsys fs.FS) ([]ClosedSeam, error) {
	ents, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read the closed-seam fragments: %w", err)
	}
	var out []ClosedSeam
	titles := map[string]string{}
	legacy := map[int]string{}
	for _, e := range ents {
		if e.IsDir() {
			return nil, fmt.Errorf("closed-seam fragments: unexpected directory %s", e.Name())
		}
		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("closed-seam fragments: %w", err)
		}
		s, err := ParseClosedSeam(e.Name(), data)
		if err != nil {
			return nil, err
		}
		if other, dup := titles[s.Title]; dup {
			return nil, fmt.Errorf("closed-seam fragments %s and %s have the same title %q", other, s.File, s.Title)
		}
		titles[s.Title] = s.File
		if s.LegacyOrder > 0 {
			if other, dup := legacy[s.LegacyOrder]; dup {
				return nil, fmt.Errorf("closed-seam fragments %s and %s have the same legacy_order %d", other, s.File, s.LegacyOrder)
			}
			legacy[s.LegacyOrder] = s.File
		}
		out = append(out, s)
	}
	SortClosedSeams(out)
	return out, nil
}

// SortClosedSeams puts fragments in list order: every fragment written
// since the migration first, newest date first (ties: the higher first
// issue number, then the higher PR, then the file name), then the
// migrated fragments in their legacy_order.
func SortClosedSeams(seams []ClosedSeam) {
	firstIssue := func(s ClosedSeam) int {
		if len(s.Issues) == 0 {
			return 0
		}
		return s.Issues[0]
	}
	sort.SliceStable(seams, func(i, j int) bool {
		a, b := seams[i], seams[j]
		if (a.LegacyOrder == 0) != (b.LegacyOrder == 0) {
			return a.LegacyOrder == 0
		}
		if a.LegacyOrder != 0 {
			return a.LegacyOrder < b.LegacyOrder
		}
		if a.Date != b.Date {
			return a.Date > b.Date
		}
		if ai, bi := firstIssue(a), firstIssue(b); ai != bi {
			return ai > bi
		}
		if a.PR != b.PR {
			return a.PR > b.PR
		}
		return a.File < b.File
	})
}

// ClosedSeamsBlock renders the closed-seams list, markers included, from
// fragments already in list order. Each entry is one list item; a
// multi-line body keeps its lines inside the item by indenting them.
// Deterministic: same fragments, same bytes.
func ClosedSeamsBlock(seams []ClosedSeam) string {
	var b strings.Builder
	b.WriteString(ClosedBeginMarker + "\n\n")
	for _, s := range seams {
		for i, line := range strings.Split(s.Body, "\n") {
			switch {
			case i == 0:
				b.WriteString("- " + line)
			case line == "":
			default:
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(ClosedEndMarker)
	return b.String()
}

// SpliceClosedSeams replaces the generated closed-seams block inside
// doc. ok is false when the markers are missing or out of order.
func SpliceClosedSeams(doc, block string) (string, bool) {
	start := strings.Index(doc, ClosedBeginMarker)
	end := strings.Index(doc, ClosedEndMarker)
	if start < 0 || end < start {
		return doc, false
	}
	return doc[:start] + block + doc[end+len(ClosedEndMarker):], true
}

// ExtractClosedSeams returns the generated closed-seams block currently
// in doc, markers included.
func ExtractClosedSeams(doc string) (string, bool) {
	start := strings.Index(doc, ClosedBeginMarker)
	end := strings.Index(doc, ClosedEndMarker)
	if start < 0 || end < start {
		return "", false
	}
	return doc[start : end+len(ClosedEndMarker)], true
}

// closedEntryTitles returns the bold title of every list entry in a
// rendered closed-seams block, in order.
func closedEntryTitles(block string) []string {
	var out []string
	for _, line := range strings.Split(block, "\n") {
		if !strings.HasPrefix(line, "- **") {
			continue
		}
		t := strings.TrimPrefix(line, "- **")
		if i := strings.Index(t, "**"); i >= 0 {
			out = append(out, t[:i])
		}
	}
	return out
}
