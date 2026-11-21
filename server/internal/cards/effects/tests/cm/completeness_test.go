package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// The zero value has to mean "unreviewed". Everything downstream —
// the public catalog page most of all — relies on a spec that forgot
// the field publishing as unaudited rather than as working in full.
func TestZeroCompletenessIsUnreviewed(t *testing.T) {
	var zero effects.Completeness
	if zero != effects.CompletenessUnreviewed {
		t.Fatalf("zero value = %v, want CompletenessUnreviewed", zero)
	}
	if got := (effects.Spec{}).Completeness.String(); got != "unreviewed" {
		t.Fatalf("Spec{}.Completeness = %q, want unreviewed", got)
	}
}

// The strings are the wire contract: the catalog JSON and the
// client's filter both key off them.
func TestCompletenessStrings(t *testing.T) {
	for c, want := range map[effects.Completeness]string{
		effects.CompletenessUnreviewed: "unreviewed",
		effects.CompletenessFull:       "full",
		effects.CompletenessCaveats:    "caveats",
		effects.Completeness(99):       "unreviewed",
	} {
		if got := c.String(); got != want {
			t.Errorf("Completeness(%d).String() = %q, want %q", c, got, want)
		}
	}
}

// A caveat card that does not say what the caveat is would publish a
// warning with no content. Caught at boot, not in a browser.
func TestRegisterRejectsCaveatsWithoutText(t *testing.T) {
	effectstest.MustPanic(t, "no Caveats", func() {
		effects.Register(effects.Spec{
			OracleID:     "completeness-test-no-text",
			Name:         "No Text",
			Completeness: effects.CompletenessCaveats,
		})
	})
}

// The mirror mistake: caveat text on a card that claims to be
// complete. One of the two is wrong and the registry cannot tell
// which, so it refuses both.
func TestRegisterRejectsCaveatsOnANonCaveatCard(t *testing.T) {
	effectstest.MustPanic(t, "use CompletenessCaveats", func() {
		effects.Register(effects.Spec{
			OracleID:     "completeness-test-full-with-caveats",
			Name:         "Full With Caveats",
			Completeness: effects.CompletenessFull,
			Caveats:      []string{"something"},
		})
	})
	effectstest.MustPanic(t, "use CompletenessCaveats", func() {
		effects.Register(effects.Spec{
			OracleID: "completeness-test-unreviewed-with-caveats",
			Name:     "Unreviewed With Caveats",
			Caveats:  []string{"something"},
		})
	})
}

func TestRegisterRejectsEmptyCaveatText(t *testing.T) {
	effectstest.MustPanic(t, "empty caveat", func() {
		effects.Register(effects.Spec{
			OracleID:     "completeness-test-empty-caveat",
			Name:         "Empty Caveat",
			Completeness: effects.CompletenessCaveats,
			Caveats:      []string{"real one", ""},
		})
	})
}

// An undeclared card must register without complaint. Making this a
// boot failure would push reviewers to stamp CompletenessFull to get
// a green build, which is the outcome the whole field exists to
// prevent — see completeness.go.
func TestRegisterAcceptsAnUndeclaredSpec(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "completeness-test-undeclared",
		Name:     "Undeclared",
	})
	got, ok := effects.Lookup("completeness-test-undeclared")
	if !ok {
		t.Fatal("spec did not register")
	}
	if got.Completeness != effects.CompletenessUnreviewed {
		t.Fatalf("completeness = %v, want unreviewed", got.Completeness)
	}
}

// Every caveat the catalog publishes is read by a player deciding
// whether to sleeve the card, so it has to be a sentence rather than
// a keyword, and it must not leak the engine vocabulary the file's
// doc comment is for. The rule itself is PlayerFacingProblems, which
// the public roadmap's summaries share.
func TestDeclaredCaveatsReadAsPlayerFacingSentences(t *testing.T) {
	for _, s := range effects.All() {
		if s.Completeness != effects.CompletenessCaveats {
			continue
		}
		for _, cv := range s.Caveats {
			for _, p := range effects.PlayerFacingProblems(cv) {
				t.Errorf("%s: caveat %s: %q", s.Name, p, cv)
			}
		}
	}
}

func TestPlayerFacingProblemsNamesEachFailure(t *testing.T) {
	if got := effects.PlayerFacingProblems("Flashback isn't implemented yet."); got != nil {
		t.Fatalf("a good sentence was refused: %v", got)
	}
	got := effects.PlayerFacingProblems("OnResolve skips it")
	if len(got) != 3 {
		t.Fatalf("want terse, no full stop and jargon; got %v", got)
	}
}
