package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_mind_stone_test.go — #1321's proof card. The ∞ trigger is the
// point: it does not exist at all until CR 701.64a's "Harness"
// activated ability has set the CR 701.64 designation (ADR 0071
// amendment) — not a condition checked at resolution, an ABSENCE from
// the harvester until then.

const theMindStoneOracle = "b175e826-09e8-4fae-9f2e-b902f95b282d"

func pushTheMindStone(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Mind Stone", OracleID: theMindStoneOracle,
		TypeLine: "Legendary Artifact — Infinity Stone", ManaCost: "{1}{W}",
		Owner: owner, Controller: owner,
	})
}

// TestTheMindStoneInfinityTriggerDoesNotFireUntilHarnessed pins the
// absence: an unharnessed Mind Stone's end step raises no prompt at
// all, because TriggersForCard drops a gated ability whose designation
// is unsatisfied before the harvester ever sees it (ADR 0071).
func TestTheMindStoneInfinityTriggerDoesNotFireUntilHarnessed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTheMindStone(g, me.ID)
	pushVanillaCreature(g, me.ID, "Untouched Bear", 2, 2)

	advanceToEndStepOf(t, g, 0)

	if pick := latestPickTarget(g, me.ID); pick != nil {
		t.Fatal("the ∞ trigger fired on an unharnessed Mind Stone")
	}
}

// TestTheMindStoneHarnessGatesTheInfinityTrigger is the whole seam,
// end to end: activating "Harness The Mind Stone" sets the
// designation, and only THEN does "∞ — at the beginning of your end
// step, exile up to one other target nonland permanent you control,
// then return it" exist to fire — as a real flicker (a new object on
// the battlefield), not a resolution-time no-op.
func TestTheMindStoneHarnessGatesTheInfinityTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushTheMindStone(g, me.ID)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	toMainForCost(t, g)

	preHarnessed := false
	g.ReadSnapshot(func() { preHarnessed = g.IsHarnessed(stone) })
	if preHarnessed {
		t.Fatal("harnessed before the ability was ever activated")
	}
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Harness: %v", err)
	}
	passPriorityAroundTable(t, g)
	harnessed := false
	g.ReadSnapshot(func() { harnessed = g.IsHarnessed(stone) })
	if !harnessed {
		t.Fatal("The Mind Stone is not harnessed after activating Harness")
	}

	advanceToEndStepOf(t, g, 0)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt at your end step once harnessed: %+v", g.PendingChoices)
	}
	if hasID(pick.PickTargetCards, stone) {
		t.Error("The Mind Stone is offered to its own \"other target nonland permanent\"")
	}
	if !hasID(pick.PickTargetCards, mine) {
		t.Fatal("a nonland permanent you control is not offered")
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)

	back := findBattlefieldByName(g, "My Bear")
	if back == uuid.Nil || back == mine {
		t.Fatalf("blinked permanent back=%v (old %v): want a new object on the battlefield", back, mine)
	}
	if c, _ := aangCardOnBF(g, back); c.Controller != me.ID {
		t.Errorf("returned under %v's control, want its owner's (%v)", c.Controller, me.ID)
	}
}

// TestTheMindStoneHarnessIsIdempotent — CR 701.64a lets the ability be
// activated a second time; nothing stops it, and the second
// activation simply does nothing (the designation is already set).
func TestTheMindStoneHarnessIsIdempotent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushTheMindStone(g, me.ID)
	toMainForCost(t, g)

	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate Harness (%d): %v", i, err)
		}
		passPriorityAroundTable(t, g)
		// The {T} half of the cost taps it every time; untap between
		// activations so the SECOND one is testing idempotency and
		// not just failing on an already-tapped source.
		if err := g.TapCard(stone, false); err != nil {
			t.Fatalf("untap for the next activation: %v", err)
		}
	}
	harnessed := false
	g.ReadSnapshot(func() { harnessed = g.IsHarnessed(stone) })
	if !harnessed {
		t.Fatal("not harnessed after two activations")
	}
}

// TestTheMindStoneIsIndestructibleAndTapsForWhite is the rest of the
// printed card — no simplification claimed, so it should not need
// one.
func TestTheMindStoneIsIndestructibleAndTapsForWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushTheMindStone(g, me.ID)

	if !hasString(effectiveAbilities(t, g, stone), "indestructible") {
		t.Error("The Mind Stone is not indestructible")
	}

	abilities := game.ManaAbilitiesForCard(battlefieldCardCopy(t, g, stone))
	if len(abilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(abilities))
	}
}
