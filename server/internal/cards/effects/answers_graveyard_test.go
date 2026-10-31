package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_graveyard_test.go — ADR 0009's amendment of 2026-10-30 (#4734),
// owner decision Q5 (a), slice 9: graveyard exile is its own answer kind,
// `graveyard`, split from `restrict` the way `survive` was split from
// `protect`. The eight rows that exile from graveyards were `restrict`
// under ADR 0142 sweep ruling 9; they declare `graveyard` now, and
// `restrict` keeps what stops an opponent doing something (Ranger-Captain
// of Eos, a turn-off-protection row, a lock lifter).

// TestGraveyardExileRowsDeclareGraveyard pins the eight re-declared rows:
// the three that exile with no target, the one that targets players, and
// the four built by ExileFromASingleGraveyardAbility.
func TestGraveyardExileRowsDeclareGraveyard(t *testing.T) {
	rows := []struct{ card, label string }{
		{"Soul-Guide Lantern", "Exile each opponent's graveyard"},
		{"Scavenger Grounds", "Exile all graveyards"},
		{"Silent Gravestone", "all cards from all graveyards"},
		{"Stonespeaker Crystal", "target players' graveyards"},
		{"Famished Ghoul", "from a single graveyard"},
		{"Rag Dealer", "from a single graveyard"},
		{"Unlicensed Hearse", "from a single graveyard"},
		{"Carrion Beetles", "from a single graveyard"},
	}
	for _, r := range rows {
		t.Run(r.card, func(t *testing.T) {
			spec := specNamed(t, r.card)
			n := 0
			for _, ab := range spec.Activated {
				if !strings.Contains(ab.Label, r.label) {
					continue
				}
				n++
				if got := ab.Purpose.Answers; got != game.AnswerGraveyard {
					t.Errorf("%q declares %s, want graveyard alone", ab.Label, got)
				}
			}
			if n != 1 {
				t.Errorf("%d rows of %s are labelled %q, want 1", n, r.card, r.label)
			}
		})
	}
}

// TestRestrictRowsAreNotGraveyardHate: no row that declares exactly
// `restrict` exiles from a graveyard. A new graveyard-exile row that
// declares `restrict` out of habit would stop answering the reanimation
// triggers the new kind exists for (Q5). The label is read here, in a
// test, as a review aid; the reader never reads text (ADR 0142 S4).
func TestRestrictRowsAreNotGraveyardHate(t *testing.T) {
	n := 0
	for _, r := range answersRows(t) {
		if r.declared() != game.AnswerRestrict {
			continue
		}
		n++
		if strings.Contains(strings.ToLower(r.label), "graveyard") {
			t.Errorf("%s (%s) declares restrict but its label is graveyard hate; declare graveyard (ADR 0009 amendment of 2026-10-30, Q5): %s", r.name, r.ref, r.label)
		}
	}
	if n == 0 {
		t.Fatal("no row declares restrict alone: the test is reading the wrong rows")
	}
}

// TestUnionRowsDeclareGraveyard: a row that declares the union of every
// answer a spell can give (Isochron Scepter's imprinted instant, Baron
// Helmut Zemo's copies) includes graveyard beside restrict, since an
// instant such as Surgical Extraction exiles from a graveyard.
func TestUnionRowsDeclareGraveyard(t *testing.T) {
	for _, name := range []string{"Isochron Scepter", "Baron Helmut Zemo"} {
		spec := specNamed(t, name)
		found := false
		for _, ab := range spec.Activated {
			if a := ab.Purpose.Answers; a.HasAny(game.AnswerRestrict) {
				found = true
				if !a.HasAny(game.AnswerGraveyard) {
					t.Errorf("%s declares %s: the union lacks graveyard", name, a)
				}
			}
		}
		if !found {
			t.Errorf("%s has no row declaring the union", name)
		}
	}
}
