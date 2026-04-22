package game

import (
	"testing"

	"github.com/google/uuid"
)

// layers_test.go covers the S16 sub-PR 1 layer-engine skeleton.
// Sub-PR 1 ships the types + the no-op recompute pass, so the
// observable behaviour is "nothing changed": Card.Effective()
// returns the printed characteristic verbatim, and ReadSnapshot's
// fast-path skips the recompute when the version counters match.
//
// Sub-PR 3 onwards adds real static-ability behaviour; the tests
// here are the regression guard that the no-op pass stays a no-op.

// TestEffectiveMatchesPrintedWhenNoStatics is the headline sub-PR 1
// regression: with no static abilities active (catalog has none in
// this sub-PR), Effective() must return exactly the printed values.
func TestEffectiveMatchesPrintedWhenNoStatics(t *testing.T) {
	c := Card{
		Name:      "Grizzly Bears",
		TypeLine:  "Creature — Bear",
		Power:     2,
		Toughness: 2,
	}
	eff := c.Effective()
	if eff.Power != 2 || eff.Toughness != 2 {
		t.Errorf("effective P/T = %d/%d, want 2/2", eff.Power, eff.Toughness)
	}
	if eff.Name != "Grizzly Bears" {
		t.Errorf("effective Name = %q, want %q", eff.Name, "Grizzly Bears")
	}
	if len(eff.Types) != 1 || eff.Types[0] != "Creature" {
		t.Errorf("effective Types = %v, want [Creature]", eff.Types)
	}
	if len(eff.Subtypes) != 1 || eff.Subtypes[0] != "Bear" {
		t.Errorf("effective Subtypes = %v, want [Bear]", eff.Subtypes)
	}
}

// TestParseTypeLineLegendaryHumanWizard exercises the type-line
// parser on the canonical Scryfall format with em-dash + supertype.
func TestParseTypeLineLegendaryHumanWizard(t *testing.T) {
	supertypes, types, subtypes := ParseTypeLine("Legendary Creature — Human Wizard")
	if len(supertypes) != 1 || supertypes[0] != "Legendary" {
		t.Errorf("supertypes = %v, want [Legendary]", supertypes)
	}
	if len(types) != 1 || types[0] != "Creature" {
		t.Errorf("types = %v, want [Creature]", types)
	}
	want := []string{"Human", "Wizard"}
	if len(subtypes) != len(want) {
		t.Fatalf("subtypes = %v, want %v", subtypes, want)
	}
	for i, w := range want {
		if subtypes[i] != w {
			t.Errorf("subtypes[%d] = %q, want %q", i, subtypes[i], w)
		}
	}
}

// TestParseTypeLineSorceryNoSubtypes covers the no-em-dash branch.
func TestParseTypeLineSorceryNoSubtypes(t *testing.T) {
	supertypes, types, subtypes := ParseTypeLine("Sorcery")
	if len(supertypes) != 0 {
		t.Errorf("supertypes = %v, want empty", supertypes)
	}
	if len(types) != 1 || types[0] != "Sorcery" {
		t.Errorf("types = %v, want [Sorcery]", types)
	}
	if len(subtypes) != 0 {
		t.Errorf("subtypes = %v, want empty", subtypes)
	}
}

// TestParseTypeLineEmpty covers placeholder demo cards.
func TestParseTypeLineEmpty(t *testing.T) {
	supertypes, types, subtypes := ParseTypeLine("")
	if supertypes != nil || types != nil || subtypes != nil {
		t.Errorf("empty type line should yield nil slices, got %v / %v / %v", supertypes, types, subtypes)
	}
}

// TestParseTypeLineBasicLand covers the "Basic Land — Forest" shape.
func TestParseTypeLineBasicLand(t *testing.T) {
	supertypes, types, subtypes := ParseTypeLine("Basic Land — Forest")
	if len(supertypes) != 1 || supertypes[0] != "Basic" {
		t.Errorf("supertypes = %v, want [Basic]", supertypes)
	}
	if len(types) != 1 || types[0] != "Land" {
		t.Errorf("types = %v, want [Land]", types)
	}
	if len(subtypes) != 1 || subtypes[0] != "Forest" {
		t.Errorf("subtypes = %v, want [Forest]", subtypes)
	}
}

// TestRecomputeLayersFastPathSkipsWhenVersionsMatch proves the
// fast-path: a snapshot read with no intervening event invokes
// recompute zero additional times.
func TestRecomputeLayersFastPathSkipsWhenVersionsMatch(t *testing.T) {
	g := newActiveGame(t)
	// First snapshot triggers an initial resolution if any version
	// bumps happened during game setup (none should — sub-PR 1 has no
	// listener that bumps yet).
	g.ReadSnapshot(func() {})
	baseline := g.LayerRecomputeCountForTest()

	// Second snapshot with no state change — count stays put.
	g.ReadSnapshot(func() {})
	if got := g.LayerRecomputeCountForTest(); got != baseline {
		t.Errorf("fast-path violated: recompute count went from %d to %d without any version bump", baseline, got)
	}
}

// TestRecomputeLayersRunsOnceAfterBump proves the version-counter
// invalidation: a single BumpLayerVersionForTest call followed by N
// snapshot reads triggers recompute exactly once (subsequent reads
// hit the fast path because the version is now resolved).
func TestRecomputeLayersRunsOnceAfterBump(t *testing.T) {
	g := newActiveGame(t)
	g.ReadSnapshot(func() {}) // settle baseline
	baseline := g.LayerRecomputeCountForTest()

	g.BumpLayerVersionForTest()
	for i := 0; i < 5; i++ {
		g.ReadSnapshot(func() {})
	}
	got := g.LayerRecomputeCountForTest()
	if got != baseline+1 {
		t.Errorf("recompute count = %d, want %d (one bump + N snapshots = one recompute)", got, baseline+1)
	}
}

// TestEffectiveOnBattlefieldUnchanged proves the wire-projection
// regression: a 2/2 creature on the battlefield comes through
// viewOfCard (via the snapshot path) at 2/2 — sub-PR 1's no-op pass
// must not perturb existing card characteristics.
//
// We check via Card.Effective() directly rather than the protocol
// view (which would pull in the protocol package and introduce a
// cycle). The protocol-level regression is covered by existing
// snapshot-shape tests that go untouched by sub-PR 1.
func TestEffectiveOnBattlefieldUnchanged(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	// Snapshot path must run without touching effective values.
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			eff := c.Effective()
			if eff.Power != 2 || eff.Toughness != 2 {
				t.Errorf("Bears post-snapshot effective = %d/%d, want 2/2", eff.Power, eff.Toughness)
			}
		}
	})
}
