package effects

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// #3855 (ADR 0068 amendment 2026-10-15): a waived payment asks the
// caster what they paid outside the pool only when something READS the
// mana spent, and "something reads it" is the card's own declaration,
// Spec.ReadsManaSpent. A reader that forgets to declare it silently goes
// back to reading a lower bound, and nothing at runtime would notice.
//
// So this scans the source. Every card file whose code (comments
// excluded) touches the mana-spent vocabulary must set ReadsManaSpent,
// and every file that sets it must touch the vocabulary, so a stale
// declaration cannot outlive the reader it was for. The vocabulary is
// every read surface the engine offers: the ManaSpent view and its
// Context / PaidCost / CastProvenance / CastCounts accessors (all named
// …Spent…), the helpers built on them (AdamantSpent,
// NoManaWasSpentToCast, …IfNoManaSpent, …PerManaFrom,
// noneCastWithManaInBatch), the sunburst keyword and converge, and
// expend, whose tally is mana spent on spells. A new accessor joins the
// pattern in the change that adds it.
var manaSpentReaderVocabulary = regexp.MustCompile(`Spent|Sunburst|Converge|PerManaFrom|WheneverYouExpend|CastWithManaIn|spellPaidWithTreasureMana`)

func TestManaSpentReadersDeclareIt(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	readers := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		src := string(raw)
		if !strings.Contains(src, "Register(Spec{") {
			continue // a helper file: the cards that call it declare
		}
		var code strings.Builder
		for _, line := range strings.Split(src, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			code.WriteString(line)
			code.WriteByte('\n')
		}
		body := code.String()
		declares := strings.Contains(body, "ReadsManaSpent:")
		// The declaration's own line names the vocabulary-free
		// constants (ReadsOwnPayment …), so it does not count as a read.
		reads := manaSpentReaderVocabulary.MatchString(regexp.MustCompile(`ReadsManaSpent:.*`).ReplaceAllString(body, ""))
		switch {
		case reads && !declares:
			t.Errorf("%s reads the mana spent but does not declare Spec.ReadsManaSpent (#3855): a waived payment would never ask what was paid", f)
		case declares && !reads:
			t.Errorf("%s declares Spec.ReadsManaSpent but reads nothing the vocabulary names: drop the declaration, or add the new accessor to manaSpentReaderVocabulary", f)
		}
		if reads {
			readers++
		}
	}
	// The catalog had 69 reader files when this landed; a scan that
	// finds almost none is a scan that stopped working.
	if readers < 40 {
		t.Fatalf("found only %d mana-spent reader files; the scan is broken", readers)
	}
}
