package roadmap

import (
	"flag"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// -update-closed, not the package's -update: a PR that edits
// registry.go runs -update for the Open seams table, and must not
// regenerate this block while it is at it, or every such PR carries a
// shared hunk at the top of the Closed list again (#1461). CI's push
// job and census-publish run it, scoped with -run.
var updateClosed = flag.Bool("update-closed", false,
	"rewrite the generated closed-seams list in docs/engine-seams.md from docs/engine-seams/closed/")

// staleClosedAction is what TestClosedSeamsAreCurrent does once the
// doc's block disagrees with the fragments. It is the census's
// three-way split (coverage/census_test.go, staleCensusActionFor) for
// the same reason: a PR that adds a fragment is expected to leave the
// block behind, and CI regenerates it on develop.
type staleClosedAction int

const (
	staleClosedRewrite staleClosedAction = iota
	staleClosedFail
	staleClosedSkip
)

// staleClosedActionFor reads the -update-closed flag and CENSUS_ENFORCE,
// the push job's switch for every block CI publishes. -update-closed
// wins; CENSUS_ENFORCE=="1" fails; anything else skips.
func staleClosedActionFor(update bool, enforce string) staleClosedAction {
	switch {
	case update:
		return staleClosedRewrite
	case enforce == "1":
		return staleClosedFail
	default:
		return staleClosedSkip
	}
}

func TestStaleClosedActionFor(t *testing.T) {
	tests := []struct {
		name    string
		update  bool
		enforce string
		want    staleClosedAction
	}{
		{"update wins over enforce unset", true, "", staleClosedRewrite},
		{"update wins over enforce=1", true, "1", staleClosedRewrite},
		{"enforce=1 fails", false, "1", staleClosedFail},
		{"unset skips (PR and local default)", false, "", staleClosedSkip},
		{"enforce=0 skips", false, "0", staleClosedSkip},
		{"garbage skips", false, "yes", staleClosedSkip},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := staleClosedActionFor(tc.update, tc.enforce); got != tc.want {
				t.Errorf("staleClosedActionFor(%v, %q) = %v, want %v", tc.update, tc.enforce, got, tc.want)
			}
		})
	}
}

// TestClosedSeamsAreCurrent is the drift guard for the generated
// Closed seams list.
//
// Two checks run everywhere, PR included, because each catches a
// mistake CI's refresh would otherwise erase without a word:
//
//   - every entry in the doc's block must come from a fragment. An
//     entry added to the list by hand, the habit this replaces, has no
//     fragment, and the next refresh on develop would delete it;
//   - nothing may be listed between "## Closed seams" and the BEGIN
//     marker, for the same reason.
//
// A stale block is then judged like the census: skipped in a PR or a
// local run, failed under CENSUS_ENFORCE=1 (the push job, right after
// its own refresh), rewritten under -update-closed.
func TestClosedSeamsAreCurrent(t *testing.T) {
	raw, err := os.ReadFile(SeamsDocPath)
	if err != nil {
		t.Fatalf("read %s: %v", SeamsDocPath, err)
	}
	doc := string(raw)
	seams, err := LoadClosedSeams(os.DirFS(ClosedSeamsDir))
	if err != nil {
		t.Fatal(err)
	}
	want := ClosedSeamsBlock(seams)

	got, ok := ExtractClosedSeams(doc)
	if !ok {
		t.Fatalf("%s has lost its generated closed-seams markers, so nothing lists the closed seams.\n"+
			"Put the two marker lines back under \"## Closed seams\" and run:\n\n"+
			"  go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed", SeamsDocPath)
	}

	if problems := handWrittenClosedEntries(doc, seams); len(problems) > 0 {
		t.Fatalf("%s's Closed seams section has entries that did not come from a fragment:\n\n  %s\n\n"+
			"The list is generated from docs/engine-seams/closed/*.md and CI regenerates it on develop, "+
			"so an entry written into the list by hand is deleted by the next refresh. Put each in its own "+
			"fragment instead (see the section intro) and drop it from the list.",
			SeamsDocPath, strings.Join(problems, "\n  "))
	}
	if got == want {
		return
	}

	switch staleClosedActionFor(*updateClosed, os.Getenv("CENSUS_ENFORCE")) {
	case staleClosedRewrite:
		next, ok := SpliceClosedSeams(doc, want)
		if !ok {
			t.Fatalf("markers vanished between read and splice in %s", SeamsDocPath)
		}
		if err := os.WriteFile(SeamsDocPath, []byte(next), 0o644); err != nil {
			t.Fatalf("write %s: %v", SeamsDocPath, err)
		}
		t.Logf("updated the closed-seams list in %s", SeamsDocPath)
	case staleClosedSkip:
		t.Skipf("the closed-seams list in %s is behind docs/engine-seams/closed/.\n\n"+
			"That is expected in a PR that adds a fragment: CI regenerates the list on every push to "+
			"develop and main (#1461), so do not regenerate it here; that is the shared hunk the "+
			"fragments exist to remove. Set CENSUS_ENFORCE=1 to enforce freshness, or run\n\n"+
			"  go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed\n\n"+
			"to look at the result locally.", SeamsDocPath)
	default:
		t.Errorf("the closed-seams list in %s disagrees with docs/engine-seams/closed/ after CI's own refresh "+
			"ran against this tree, so the refresh did not take. Regenerate with:\n\n"+
			"  go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed", SeamsDocPath)
	}
}

// The real fragment directory parses, and the migrated archive is
// intact: exactly LegacyClosedSeams fragments carry legacy_order and
// they are 1..LegacyClosedSeams with no gap. A gap means a migrated
// entry was deleted, and its neighbours' "the row above" now points at
// the wrong row.
func TestClosedSeamFragmentsAreWellFormed(t *testing.T) {
	seams, err := LoadClosedSeams(os.DirFS(ClosedSeamsDir))
	if err != nil {
		t.Fatal(err)
	}
	var legacy []int
	for _, s := range seams {
		if s.LegacyOrder > 0 {
			legacy = append(legacy, s.LegacyOrder)
		}
	}
	if len(legacy) != LegacyClosedSeams {
		t.Fatalf("%d fragments carry legacy_order, want %d", len(legacy), LegacyClosedSeams)
	}
	for i, n := range legacy {
		if n != i+1 {
			t.Fatalf("legacy_order runs %v…; want 1..%d with no gap", legacy[:i+1], LegacyClosedSeams)
		}
	}
}

func fragment(front, body string) string {
	return "---\n" + front + "\n---\n" + body + "\n"
}

func TestParseClosedSeam(t *testing.T) {
	s, err := ParseClosedSeam("1199-phasing.md", []byte(fragment(
		`title: "Phasing: \"in\" and out"`+"\ndate: 2026-09-23\nissues: [1199, 1200]\npr: 1251",
		"\n**Phasing: \"in\" and out** (#1199) — text.\n")))
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != `Phasing: "in" and out` || s.Date != "2026-09-23" || s.PR != 1251 || s.LegacyOrder != 0 ||
		len(s.Issues) != 2 || s.Issues[0] != 1199 || s.Issues[1] != 1200 {
		t.Errorf("parsed %+v", s)
	}
	if s.Body != `**Phasing: "in" and out** (#1199) — text.` {
		t.Errorf("body %q: the blank lines around it are trimmed, nothing else", s.Body)
	}

	none, err := ParseClosedSeam("eventattack.md", []byte(fragment(
		"title: \"EventAttack\"\ndate: 2026-09-15\nissues: []\nlegacy_order: 116", "**EventAttack** - x")))
	if err != nil {
		t.Fatal(err)
	}
	if len(none.Issues) != 0 || none.PR != 0 || none.LegacyOrder != 116 {
		t.Errorf("parsed %+v", none)
	}
}

// A malformed fragment fails loudly, and the error names the file.
func TestParseClosedSeamRejectsMalformedFragments(t *testing.T) {
	const good = "title: \"T\"\ndate: 2026-09-24\nissues: [1]"
	cases := []struct {
		name, file, text, wantErr string
	}{
		{"no front matter", "1-t.md", "**T** x\n", "must start with a front-matter line"},
		{"unclosed front matter", "1-t.md", "---\n" + good + "\n**T** x\n", "no closing"},
		{"missing title", "1-t.md", fragment("date: 2026-09-24\nissues: [1]", "**T** x"), `missing "title"`},
		{"missing date", "1-t.md", fragment("title: \"T\"\nissues: [1]", "**T** x"), `missing "date"`},
		{"missing issues", "1-t.md", fragment("title: \"T\"\ndate: 2026-09-24", "**T** x"), `missing "issues"`},
		{"unquoted title", "1-t.md", fragment("title: T\ndate: 2026-09-24\nissues: [1]", "**T** x"), "double-quoted"},
		{"title with bold", "1-t.md", fragment("title: \"a**b\"\ndate: 2026-09-24\nissues: [1]", "**a**b** x"), "must not contain"},
		{"bad date", "1-t.md", fragment("title: \"T\"\ndate: 24/09/2026\nissues: [1]", "**T** x"), "YYYY-MM-DD"},
		{"impossible date", "1-t.md", fragment("title: \"T\"\ndate: 2026-02-30\nissues: [1]", "**T** x"), "not a calendar date"},
		{"bare issue", "1-t.md", fragment("title: \"T\"\ndate: 2026-09-24\nissues: 1", "**T** x"), "list of issue numbers"},
		{"hash issue", "1-t.md", fragment("title: \"T\"\ndate: 2026-09-24\nissues: [#1]", "**T** x"), "list of issue numbers"},
		{"zero issue", "1-t.md", fragment("title: \"T\"\ndate: 2026-09-24\nissues: [0]", "**T** x"), "positive"},
		{"bad pr", "1-t.md", fragment(good+"\npr: #12", "**T** x"), "pull-request number"},
		{"unknown key", "1-t.md", fragment(good+"\nauthor: me", "**T** x"), "unknown front-matter key"},
		{"duplicate key", "1-t.md", fragment(good+"\ndate: 2026-09-25", "**T** x"), "appears twice"},
		{"not key value", "1-t.md", fragment(good+"\njust text", "**T** x"), "not \"key: value\""},
		{"legacy order on a new fragment", "1-t.md", fragment(good+"\nlegacy_order: 135", "**T** x"), "must not set it"},
		{"empty body", "1-t.md", fragment(good, "\n\n"), "no entry text"},
		{"body with list marker", "1-t.md", fragment(good, "- **T** x"), "begin with its title in bold"},
		{"body title differs", "1-t.md", fragment(good, "**Other** x"), "begin with its title in bold"},
		{"body with a marker", "1-t.md", fragment(good, "**T** <!-- END GENERATED CLOSED SEAMS -->"), "generated-block marker"},
		{"bad file name", "Phasing.md", fragment(good, "**T** x"), "file name"},
		{"not markdown", "1-t.txt", fragment(good, "**T** x"), "file name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseClosedSeam(tc.file, []byte(tc.text))
			if err == nil {
				t.Fatalf("parsed without error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) || !strings.Contains(err.Error(), tc.file) {
				t.Errorf("error %q, want it to name %s and say %q", err, tc.file, tc.wantErr)
			}
		})
	}
}

func TestLoadClosedSeamsOrder(t *testing.T) {
	frag := func(title, date, issues, extra string) *fstest.MapFile {
		front := "title: \"" + title + "\"\ndate: " + date + "\nissues: " + issues
		if extra != "" {
			front += "\n" + extra
		}
		return &fstest.MapFile{Data: []byte(fragment(front, "**"+title+"** body"))}
	}
	fsys := fstest.MapFS{
		"1-legacy-two.md":   frag("Legacy two", "2026-09-24", "[1]", "legacy_order: 2"),
		"2-legacy-one.md":   frag("Legacy one", "2026-09-10", "[2]", "legacy_order: 1"),
		"10-old.md":         frag("Old", "2026-09-25", "[10]", ""),
		"20-new.md":         frag("New", "2026-09-26", "[20]", ""),
		"30-same-day-a.md":  frag("Same day, higher issue", "2026-09-25", "[30]", ""),
		"5-same-day-pr.md":  frag("Same day, same issue, higher PR", "2026-09-25", "[10]", "pr: 99"),
		"no-issue-later.md": frag("No issue", "2026-09-25", "[]", ""),
	}
	seams, err := LoadClosedSeams(fsys)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range seams {
		got = append(got, s.Title)
	}
	want := []string{
		"New",                             // newest date first
		"Same day, higher issue",          // same date: higher first issue
		"Same day, same issue, higher PR", // same date and issue: higher PR
		"Old",
		"No issue",   // no issue sorts as 0
		"Legacy one", // the migrated archive, in legacy_order, whatever its dates
		"Legacy two",
	}
	if strings.Join(got, " | ") != strings.Join(want, " | ") {
		t.Errorf("order\n got %q\nwant %q", got, want)
	}
}

func TestLoadClosedSeamsRejectsABadDirectory(t *testing.T) {
	ok := func(title string, extra string) *fstest.MapFile {
		front := "title: \"" + title + "\"\ndate: 2026-09-24\nissues: [1]"
		if extra != "" {
			front += "\n" + extra
		}
		return &fstest.MapFile{Data: []byte(fragment(front, "**"+title+"** x"))}
	}
	cases := []struct {
		name    string
		fsys    fstest.MapFS
		wantErr string
	}{
		{"malformed fragment", fstest.MapFS{"1-a.md": ok("A", ""), "2-b.md": {Data: []byte("**B** x\n")}}, "2-b.md"},
		{"stray file", fstest.MapFS{"1-a.md": ok("A", ""), "README.txt": {Data: []byte("hi")}}, "README.txt"},
		{"subdirectory", fstest.MapFS{"1-a.md": ok("A", ""), "sub/2-b.md": ok("B", "")}, "unexpected directory"},
		{"same title twice", fstest.MapFS{"1-a.md": ok("A", ""), "2-a.md": ok("A", "")}, "same title"},
		{"same legacy_order twice", fstest.MapFS{"1-a.md": ok("A", "legacy_order: 3"), "2-b.md": ok("B", "legacy_order: 3")}, "same legacy_order"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadClosedSeams(tc.fsys)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want one mentioning %q", err, tc.wantErr)
			}
		})
	}
}

func TestClosedSeamsBlockRendering(t *testing.T) {
	block := ClosedSeamsBlock([]ClosedSeam{
		{Title: "A", Body: "**A** one line."},
		{Title: "B", Body: "**B** first line\nsecond line\n\nsecond paragraph"},
	})
	want := ClosedBeginMarker + "\n\n" +
		"- **A** one line.\n\n" +
		"- **B** first line\n  second line\n\n  second paragraph\n\n" +
		ClosedEndMarker
	if block != want {
		t.Errorf("block\n%s\nwant\n%s", block, want)
	}
	if got := closedEntryTitles(block); strings.Join(got, ",") != "A,B" {
		t.Errorf("titles %q", got)
	}

	doc := "# x\n\n## Closed seams\n\n" + ClosedBeginMarker + "\nold\n" + ClosedEndMarker + "\n"
	next, ok := SpliceClosedSeams(doc, block)
	if !ok {
		t.Fatal("splice failed")
	}
	if got, _ := ExtractClosedSeams(next); got != block {
		t.Errorf("extract after splice = %q", got)
	}
	if !strings.HasSuffix(next, ClosedEndMarker+"\n") || !strings.HasPrefix(next, "# x\n\n## Closed seams\n\n") {
		t.Errorf("splice touched text outside the markers: %q", next)
	}
	if _, ok := SpliceClosedSeams("no markers", block); ok {
		t.Error("splice without markers reported ok")
	}
}

// An entry typed into the list, or above it, has no fragment, and CI's
// next refresh would delete it; the drift test fails on it in a PR too.
func TestHandWrittenClosedEntries(t *testing.T) {
	seams := []ClosedSeam{{Title: "From a fragment", Body: "**From a fragment** x"}}
	block := ClosedSeamsBlock(seams)
	doc := func(above, inBlock string) string {
		b := block
		if inBlock != "" {
			b = strings.Replace(b, ClosedBeginMarker+"\n\n", ClosedBeginMarker+"\n\n"+inBlock+"\n\n", 1)
		}
		return "# Engine seams\n\n## Closed seams\n\nIntro.\n\n" + above + b + "\n"
	}
	if got := handWrittenClosedEntries(doc("", ""), seams); len(got) != 0 {
		t.Errorf("a generated doc reported %q", got)
	}
	if got := handWrittenClosedEntries(doc("", "- **Typed into the list** by hand"), seams); len(got) != 1 ||
		!strings.Contains(got[0], "Typed into the list") {
		t.Errorf("an entry added inside the block: %q", got)
	}
	if got := handWrittenClosedEntries(doc("- **Above the block** by hand\n\n", ""), seams); len(got) != 1 ||
		!strings.Contains(got[0], "above the generated block") {
		t.Errorf("an entry added above the block: %q", got)
	}
	if got := handWrittenClosedEntries("## Other\n\n"+block+"\n", seams); len(got) != 1 {
		t.Errorf("a block with no Closed seams heading: %q", got)
	}
}
