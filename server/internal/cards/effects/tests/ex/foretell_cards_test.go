package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Each converted card declares foretell, at the cost its oracle text
// prints, and the {2} the KEYWORD charges rather than a repeat of the
// foretell cost.
func TestForetellCardsDeclareTheirPrintedCost(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Saw It Coming", effectstest.SawItComingOracle, "{1}{U}"},
		{"Behold the Multiverse", effectstest.BeholdTheMultiverseOracl, "{1}{U}"},
		{"Cosmic Intervention", effectstest.CosmicInterventionOracle, "{1}{W}"},
		{"Spectral Deluge", effectstest.B41SpectralDelugeOracle, "{1}{U}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := effectstest.ForetellOf(tc.oracle)
			if sa == nil {
				t.Fatalf("%s declares no foretell special action", tc.name)
			}
			if sa.CastCost != tc.cost {
				t.Errorf("foretell cost: got %q, want %q", sa.CastCost, tc.cost)
			}
			if sa.Cost != game.ForetellExileCost {
				t.Errorf("special-action cost: got %q, want the keyword's %q", sa.Cost, game.ForetellExileCost)
			}
		})
	}
}

// End to end on the real card: foretell Saw It Coming on your turn,
// then counter a spell with it on your next turn for {1}{U} — two
// mana that were never held up.
func TestSawItComingIsForetoldAndCastLaterForItsForetellCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(active, "Saw It Coming", "Instant", effectstest.SawItComingOracle)

	active.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(active.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell Saw It Coming: %v", err)
	}

	var inExile game.Card
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if ok {
			inExile = c
		}
	})
	if !game.CardIsForetold(inExile) {
		t.Fatalf("Saw It Coming is not a foretold card in exile: FaceDown=%v kind=%q",
			inExile.FaceDown, inExile.FaceDownKind)
	}

	// A later turn. The permission's floor is a turn NUMBER, which
	// counts rounds, so this is the seat's next turn.
	g.WithWriteLock(func() { g.Turn.Seq++ })

	// Something to counter.
	victim := effectstest.CastCatalogSpell(t, g, "Filler Bolt", "Instant", "test-foretell-victim", nil)

	active.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if err := g.CastSpell(active.ID, card, game.CastSpellParams{
		Strict:          true,
		FromZone:        "exile",
		AlternativeCost: game.AltCostKeyForetell,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast the foretold Saw It Coming: %v", err)
	}
	if len(active.ManaPool) != 0 {
		t.Errorf("mana pool: got %d tokens left, want 0 — the foretell cost was not {1}{U}", len(active.ManaPool))
	}
	var foretold bool
	g.ReadSnapshot(func() {
		if item := g.StackMeta[card]; item != nil {
			foretold = item.Foretold
		}
	})
	if !foretold {
		t.Error("the spell is not marked foretold on the stack (CR 702.143c)")
	}
}

// Behold the Multiverse: the printed cast still scries and draws, and
// foretell does not interfere with either.
func TestBeholdTheMultiverseScriesAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	before := active.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Behold the Multiverse", "Instant", effectstest.BeholdTheMultiverseOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("Behold the Multiverse queued no scry prompt")
	}
	if scry.Count != 2 {
		t.Errorf("scry count: got %d, want 2", scry.Count)
	}
	if err := g.ResolveScry(scry.ID, active.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	// castCatalogSpell seeds the spell into the hand and then casts
	// it, so the cast itself is hand-neutral and the draw is the whole
	// delta.
	if got, want := active.Hand.Size(), before+2; got != want {
		t.Errorf("hand size: got %d, want %d (drew two)", got, want)
	}
}

// Cosmic Intervention's caveats no longer claim foretell is missing.
// Its last caveat (the dying commander) went with #1306, so the card
// is full; the foretell check stays as the guard against the old
// claim coming back.
func TestCosmicInterventionCaveatsDropTheForetellClaim(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.CosmicInterventionOracle)
	if !ok {
		t.Fatal("Cosmic Intervention is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Cosmic Intervention completeness = %v, want full", spec.Completeness)
	}
	for _, cv := range spec.Caveats {
		if effectstest.ContainsFold(cv, "foretell") {
			t.Errorf("Cosmic Intervention still caveats foretell: %q", cv)
		}
	}
}
