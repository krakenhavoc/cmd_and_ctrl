package effects

import (
	"strings"
	"testing"
)

// The zero value has to mean "unreviewed". Everything downstream —
// the public catalog page most of all — relies on a spec that forgot
// the field publishing as unaudited rather than as working in full.
func TestZeroCompletenessIsUnreviewed(t *testing.T) {
	var zero Completeness
	if zero != CompletenessUnreviewed {
		t.Fatalf("zero value = %v, want CompletenessUnreviewed", zero)
	}
	if got := (Spec{}).Completeness.String(); got != "unreviewed" {
		t.Fatalf("Spec{}.Completeness = %q, want unreviewed", got)
	}
}

// The strings are the wire contract: the catalog JSON and the
// client's filter both key off them.
func TestCompletenessStrings(t *testing.T) {
	for c, want := range map[Completeness]string{
		CompletenessUnreviewed: "unreviewed",
		CompletenessFull:       "full",
		CompletenessCaveats:    "caveats",
		Completeness(99):       "unreviewed",
	} {
		if got := c.String(); got != want {
			t.Errorf("Completeness(%d).String() = %q, want %q", c, got, want)
		}
	}
}

func mustPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic; wanted one mentioning %q", want)
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic = %v, want one mentioning %q", r, want)
		}
	}()
	fn()
}

// A caveat card that does not say what the caveat is would publish a
// warning with no content. Caught at boot, not in a browser.
func TestRegisterRejectsCaveatsWithoutText(t *testing.T) {
	mustPanic(t, "no Caveats", func() {
		Register(Spec{
			OracleID:     "completeness-test-no-text",
			Name:         "No Text",
			Completeness: CompletenessCaveats,
		})
	})
}

// The mirror mistake: caveat text on a card that claims to be
// complete. One of the two is wrong and the registry cannot tell
// which, so it refuses both.
func TestRegisterRejectsCaveatsOnANonCaveatCard(t *testing.T) {
	mustPanic(t, "use CompletenessCaveats", func() {
		Register(Spec{
			OracleID:     "completeness-test-full-with-caveats",
			Name:         "Full With Caveats",
			Completeness: CompletenessFull,
			Caveats:      []string{"something"},
		})
	})
	mustPanic(t, "use CompletenessCaveats", func() {
		Register(Spec{
			OracleID: "completeness-test-unreviewed-with-caveats",
			Name:     "Unreviewed With Caveats",
			Caveats:  []string{"something"},
		})
	})
}

func TestRegisterRejectsEmptyCaveatText(t *testing.T) {
	mustPanic(t, "empty caveat", func() {
		Register(Spec{
			OracleID:     "completeness-test-empty-caveat",
			Name:         "Empty Caveat",
			Completeness: CompletenessCaveats,
			Caveats:      []string{"real one", ""},
		})
	})
}

// An undeclared card must register without complaint. Making this a
// boot failure would push reviewers to stamp CompletenessFull to get
// a green build, which is the outcome the whole field exists to
// prevent — see completeness.go.
func TestRegisterAcceptsAnUndeclaredSpec(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: "completeness-test-undeclared",
		Name:     "Undeclared",
	})
	got, ok := Lookup("completeness-test-undeclared")
	if !ok {
		t.Fatal("spec did not register")
	}
	if got.Completeness != CompletenessUnreviewed {
		t.Fatalf("completeness = %v, want unreviewed", got.Completeness)
	}
}

// Every caveat the catalog publishes is read by a player deciding
// whether to sleeve the card, so it has to be a sentence rather than
// a keyword, and it must not leak the engine vocabulary the file's
// doc comment is for.
func TestDeclaredCaveatsReadAsPlayerFacingSentences(t *testing.T) {
	jargon := []string{
		"OnResolve", "OnETB", "AsEnters", "TargetSpec", "InstanceID", "StackItem",
		"ctx.", "*Game", "Locked", "sub-PR",
	}
	for _, s := range All() {
		if s.Completeness != CompletenessCaveats {
			continue
		}
		for _, cv := range s.Caveats {
			if len(cv) < 20 {
				t.Errorf("%s: caveat too terse to be useful: %q", s.Name, cv)
			}
			if !strings.HasSuffix(cv, ".") {
				t.Errorf("%s: caveat is not a sentence: %q", s.Name, cv)
			}
			for _, j := range jargon {
				if strings.Contains(cv, j) {
					t.Errorf("%s: caveat leaks engine jargon %q: %q", s.Name, j, cv)
				}
			}
		}
	}
}
