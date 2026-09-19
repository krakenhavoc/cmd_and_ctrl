package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foretell_cards_test.go — #658, the three cards, against the REAL
// catalog. The model and the verb are pinned in the game package;
// what only a catalog test can check is that each card file actually
// declared foretell, and declared it at the printed cost.

const (
	sawItComingOracle        = "90edaf33-d0ab-47e0-8f6a-6fba38286e6e"
	beholdTheMultiverseOracl = "d7d2f701-77df-4169-bf98-0d51d6886e9b"
)

// foretellOf reads the card's declared foretell action, or nil.
func foretellOf(oracle string) *game.SpecialAction {
	for _, sa := range game.SpecialActionsFor(oracle) {
		if sa.Kind == game.SpecialActionForetell {
			out := sa
			return &out
		}
	}
	return nil
}

// Each converted card declares foretell, at the cost its oracle text
// prints, and the {2} the KEYWORD charges rather than a repeat of the
// foretell cost.
func TestForetellCardsDeclareTheirPrintedCost(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Saw It Coming", sawItComingOracle, "{1}{U}"},
		{"Behold the Multiverse", beholdTheMultiverseOracl, "{1}{U}"},
		{"Cosmic Intervention", cosmicInterventionOracle, "{1}{W}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := foretellOf(tc.oracle)
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
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(active, "Saw It Coming", "Instant", sawItComingOracle)

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
	g.WithWriteLock(func() { g.Turn.Number++ })

	// Something to counter.
	victim := castCatalogSpell(t, g, "Filler Bolt", "Instant", "test-foretell-victim", nil)

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
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	before := active.Hand.Size()
	castCatalogSpell(t, g, "Behold the Multiverse", "Instant", beholdTheMultiverseOracl, nil)
	passPriorityAroundTable(t, g)

	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
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

// Cosmic Intervention's caveat no longer claims foretell is missing,
// and the card still says what IS missing.
func TestCosmicInterventionCaveatsDropTheForetellClaim(t *testing.T) {
	spec, ok := Lookup(cosmicInterventionOracle)
	if !ok {
		t.Fatal("Cosmic Intervention is not registered")
	}
	if len(spec.Caveats) == 0 {
		t.Fatal("Cosmic Intervention declares no caveats at all")
	}
	for _, cv := range spec.Caveats {
		if containsFold(cv, "foretell") {
			t.Errorf("Cosmic Intervention still caveats foretell: %q", cv)
		}
	}
}

// containsFold is a case-insensitive substring test, kept local so the
// assertion above reads as one line.
func containsFold(haystack, needle string) bool {
	h, n := []rune(haystack), []rune(needle)
	lower := func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}
	if len(n) == 0 || len(n) > len(h) {
		return false
	}
	for i := 0; i+len(n) <= len(h); i++ {
		match := true
		for j := range n {
			if lower(h[i+j]) != lower(n[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
