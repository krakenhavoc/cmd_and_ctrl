package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspend_cards_test.go — #659, the three cards, against the REAL
// catalog. The mechanic is pinned in the game package; what only a
// catalog test can check is that each card file declared suspend at
// its printed N and cost, that the countdown trigger the KEYWORD owns
// really is attached to each of them, and that the two cards with no
// mana cost are playable at all.

const (
	riftBoltOracle        = "2b8afa9f-4236-4c02-a8d5-3c145caecfd6"
	lotusBloomOracle      = "04cf02dc-f053-414e-87d8-1537f25bcbf4"
	ancestralVisionOracle = "9728dec9-d482-4c7a-8cdc-44d010dc878d"
)

// Each converted card declares suspend at its printed N and cost.
func TestSuspendCardsDeclareTheirPrintedNumbers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		n      int
		cost   string
	}{
		{"Rift Bolt", riftBoltOracle, 1, "{R}"},
		{"Lotus Bloom", lotusBloomOracle, 3, "{0}"},
		{"Ancestral Vision", ancestralVisionOracle, 4, "{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := effectstest.SuspendOf(tc.oracle)
			if sa == nil {
				t.Fatalf("%s declares no suspend special action", tc.name)
			}
			if sa.Counters != tc.n {
				t.Errorf("time counters: got %d, want %d", sa.Counters, tc.n)
			}
			if sa.Cost != tc.cost {
				t.Errorf("suspend cost: got %q, want %q", sa.Cost, tc.cost)
			}
		})
	}
}

// Both triggers belong to the KEYWORD, not to the card file: buildDef
// grows them from the declaration, so every suspend card has exactly
// the two CR 702.62b prints — the upkeep countdown and "when the last
// time counter is removed" (#990) — and none of them wrote either out.
func TestEverySuspendCardCarriesBothCountdownTriggers(t *testing.T) {
	for _, oracle := range []string{riftBoltOracle, lotusBloomOracle, ancestralVisionOracle} {
		var upkeep, lastCounter int
		for _, tr := range game.CatalogTriggers(oracle) {
			if !game.TriggerWatchesFromZone(tr, game.ZoneExile) {
				continue
			}
			for _, kind := range tr.Watches {
				switch kind {
				case game.EventBeginUpkeep:
					upkeep++
				case game.EventCounterPlaced:
					lastCounter++
				}
			}
		}
		if upkeep != 1 || lastCounter != 1 {
			t.Errorf("%s carries %d upkeep and %d last-counter exile triggers, want one of each (CR 702.62b)",
				oracle, upkeep, lastCounter)
		}
	}
}

// CR 118.6: a card with no mana cost cannot be cast by paying it.
// Ancestral Vision and Lotus Bloom are the two cards in the catalog
// that prove it, and suspend is the only way either is ever played
// (CR 118.6a).
func TestTheNoManaCostSuspendCardsCannotBeHardCast(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		oracle   string
	}{
		{"Ancestral Vision", "Sorcery", ancestralVisionOracle},
		{"Lotus Bloom", "Artifact", lotusBloomOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			active := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			id := effectstest.HandCardForTest(active, tc.name, tc.typeLine, tc.oracle)
			// CR 118.6 reads the Scryfall-stamped Layout as its "this
			// card's printed fields are real" guard — a hand-built
			// fixture with no layout means "unknown cost", not "no
			// cost" — so the fixture has to look imported.
			effectstest.StampLayout(g, active, id, "normal")
			err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true})
			if err == nil {
				t.Fatalf("%s was cast from hand for its (unpayable) mana cost", tc.name)
			}
		})
	}
}

// End to end on the real card: suspend Rift Bolt for {R}, tick once on
// the next upkeep, take the offer, and cast a SORCERY in an upkeep for
// nothing.
func TestRiftBoltSuspendsTicksAndCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(active, "Rift Bolt", "Sorcery", riftBoltOracle)
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})

	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend Rift Bolt: %v", err)
	}
	var counters int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			counters = c.Counters[game.CounterTime]
		}
	})
	if counters != 1 {
		t.Fatalf("Rift Bolt has %d time counters after suspending, want 1", counters)
	}

	// Walk to this seat's next upkeep and let the countdown resolve.
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)

	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}
	// #4250: the cast below answers the resolution_cast prompt.

	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := victim.Life
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Strict:   true,
		FromZone: "exile",
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("the free cast of a sorcery in an upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := victim.Life; got != before-3 {
		t.Errorf("victim life: got %d, want %d", got, before-3)
	}
}
