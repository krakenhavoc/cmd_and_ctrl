package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// additional_cost_test.go — S21 sub-PR 5: the "As an additional cost
// to cast this spell, discard a card" trio. The interesting cases
// aren't the card draw, they're the ordering: the discard is part of
// casting, so the payoffs trigger above the spell and resolve first.

const (
	thrillOfPossibilityOracle = "1cb0610b-a731-42c2-b93f-0a29f63cebf4"
	bigScoreOracle            = "a5cbd257-c836-493e-bb1a-76242619dea2"
	unexpectedWindfallOracle  = "498c10c9-253d-4b15-b48c-1509381b17e8"
)

// castWithDiscard seeds a card into the active seat's hand along
// with `discards` cards of the given type line, walks to a main
// phase, and casts the spell paying those cards. Returns the spell
// and the discarded IDs.
func castWithDiscard(t *testing.T, g *game.Game, name, oracle, discardType string, discards int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	paid := make([]uuid.UUID, 0, discards)
	for i := 0; i < discards; i++ {
		paid = append(paid, handCard(active, "Fodder", discardType))
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{DiscardIDs: paid}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id, paid
}

func TestAdditionalCostsAreWired(t *testing.T) {
	for _, oracle := range []string{thrillOfPossibilityOracle, bigScoreOracle, unexpectedWindfallOracle} {
		cost := game.AdditionalCostFor(oracle)
		if cost.Empty() {
			t.Errorf("%s: expected an additional cost", oracle)
			continue
		}
		if cost.DiscardCards != 1 {
			t.Errorf("%s: DiscardCards = %d, want 1", oracle, cost.DiscardCards)
		}
		if cost.Label != "Discard a card" {
			t.Errorf("%s: Label = %q", oracle, cost.Label)
		}
	}
	// The overwhelming majority of cards have none, and the hook
	// must say so rather than allocating an empty cost.
	if game.AdditionalCostFor(faithlessLootingOracle) != nil {
		t.Errorf("Faithless Looting has no additional cost")
	}
}

func TestThrillOfPossibilityDrawsTwoAfterTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()

	_, paid := castWithDiscard(t, g, "Thrill of Possibility", thrillOfPossibilityOracle, "Sorcery", 1)
	// Paid at announce: the card is already gone before resolution.
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatalf("the discard should be paid at announce, not on resolution")
	}
	passPriorityAroundTable(t, g)

	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (fodder discarded, spell cast, two drawn)", me.Hand.Size())
	}
}

func TestBigScoreAndWindfallDrawTwoAndMakeTwoTreasures(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Big Score", bigScoreOracle},
		{"Unexpected Windfall", unexpectedWindfallOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Hand.Cards = nil
			before := me.Library.Size()

			castWithDiscard(t, g, tc.name, tc.oracle, "Sorcery", 1)
			passPriorityAroundTable(t, g)

			if got := before - me.Library.Size(); got != 2 {
				t.Errorf("drew %d cards, want 2", got)
			}
			treasures := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Treasure" && c.Controller == me.ID {
					treasures++
				}
			}
			if treasures != 2 {
				t.Errorf("Treasures = %d, want 2", treasures)
			}
		})
	}
}

// The point of modelling the cost: Mary Read sees the discard while
// the spell is still on the stack, so her Treasure is on the
// battlefield before the spell resolves.
func TestAdditionalCostDiscardTriggersTheCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	pushCatalogPermanent(g, me.ID, "Mary Read and Anne Bonny",
		"Legendary Creature — Human Assassin Pirate", maryReadOracle, false)

	spell, _ := castWithDiscard(t, g, "Thrill of Possibility", thrillOfPossibilityOracle,
		"Creature — Pirate", 1)

	// One pass drains the trigger onto the stack and resolves it —
	// it's above the spell, so the Treasure lands first.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8 && findBattlefieldByName(g, "Treasure") == uuid.Nil; i++ {
		if !g.Stack.Contains(spell) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	treasure := findBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatalf("discarding a Pirate to the cost should have made a Treasure")
	}
	if !g.Stack.Contains(spell) {
		t.Errorf("the Treasure should arrive while the spell is still on the stack")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == treasure && !c.Tapped {
			t.Errorf("Mary Read's Treasure enters tapped")
		}
	}
}

// And the Mako grows from a cost payment exactly as it does from a
// loot.
func TestAdditionalCostDiscardGrowsTheMako(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)

	castWithDiscard(t, g, "Big Score", bigScoreOracle, "Sorcery", 1)
	passPriorityAroundTable(t, g)

	counters := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mako {
			counters = g.Battlefield.Cards[i].Counters["+1/+1"]
		}
	}
	if counters != 1 {
		t.Errorf("Mako counters = %d, want 1", counters)
	}
}

// A cast that doesn't pay the cost is rejected, and leaves the hand
// untouched.
func TestCastingWithoutPayingTheAdditionalCostIsRejected(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Thrill of Possibility", TypeLine: "Instant",
		OracleID: thrillOfPossibilityOracle, Owner: me.ID, Controller: me.ID,
	})
	handCard(me, "Fodder", "Sorcery")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != game.ErrInvalidParam {
		t.Fatalf("cast without the discard: %v, want ErrInvalidParam", err)
	}
	if me.Hand.Size() != 2 || me.Graveyard.Size() != 0 {
		t.Errorf("rejected cast changed the board: hand %d, graveyard %d", me.Hand.Size(), me.Graveyard.Size())
	}
}
