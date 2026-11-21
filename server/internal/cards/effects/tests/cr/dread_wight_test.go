package testscr

import (
	"testing"

	"github.com/google/uuid"

	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oracleDreadWight = "f9078d79-3250-44e8-aba1-ab52e0ba4425"

// dwCreature puts a creature with real P/T on the battlefield, so combat
// damage leaves both sides alive.
func dwCreature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, name, "Creature — Zombie", oracle, false)
	c := effectstest.FindBattlefieldCardForTest(g, id)
	c.Power, c.Toughness = power, toughness
	return id
}

func dwCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[cardsdi.DreadWightCounter]
}

// dwToEndOfCombat runs the active seat's `attacker` into `defender`,
// blocked by `blocker`, and resolves the end-of-combat trigger.
func dwToEndOfCombat(t *testing.T, g *game.Game, defender, attacker, blocker uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, defender, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
}

// The blocker is paralyzed: a counter, tapped, held through its
// controller's untap steps, and free again once it pays {4} to remove
// the counter. The granted ability stays after the counter is gone.
func TestDreadWightParalyzesItsBlockerUntilTheCounterIsRemoved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wight := dwCreature(g, me.ID, "Dread Wight", oracleDreadWight, 3, 4)
	blocker := dwCreature(g, opp.ID, "Wall", "", 0, 5)
	bystander := dwCreature(g, opp.ID, "Bystander", "", 2, 2)
	effectstest.FindBattlefieldCardForTest(g, bystander).Tapped = true

	dwToEndOfCombat(t, g, opp.ID, wight, blocker)
	if dwCounters(t, g, blocker) != 1 || !effectstest.TappedForTest(t, g, blocker) || effectstest.HoldCount(t, g, blocker) != 1 {
		t.Fatalf("blocker: %d counters, tapped %v, %d holds; want one counter, tapped, one hold",
			dwCounters(t, g, blocker), effectstest.TappedForTest(t, g, blocker), effectstest.HoldCount(t, g, blocker))
	}
	if dwCounters(t, g, wight) != 0 || effectstest.HoldCount(t, g, wight) != 0 {
		t.Errorf("the Wight paralyzed itself: %d counters", dwCounters(t, g, wight))
	}
	if dwCounters(t, g, bystander) != 0 || effectstest.HoldCount(t, g, bystander) != 0 {
		t.Error("a creature outside the Wight's combat was paralyzed")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, blocker); idx < 0 {
		t.Fatal("the blocker did not gain the remove-a-counter ability")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.TappedForTest(t, g, blocker) {
		t.Fatal("the paralyzed creature untapped during its controller's untap step")
	}
	if effectstest.TappedForTest(t, g, bystander) {
		t.Error("the bystander stayed tapped")
	}

	idx, ref := effectstest.GrantedActivatedIndex(t, g, blocker)
	effectstest.B06AddMana(opp, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, blocker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if dwCounters(t, g, blocker) != 0 {
		t.Fatalf("%d paralyzation counters after paying {4}, want 0", dwCounters(t, g, blocker))
	}
	if effectstest.HoldCount(t, g, blocker) != 0 {
		t.Error("the hold outlived the last paralyzation counter")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, blocker); idx < 0 {
		t.Error("the granted ability ended with the counter; it prints no duration")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, blocker) {
		t.Error("the creature stayed tapped after its last paralyzation counter was removed")
	}
}

// A Wight that blocks paralyzes the attacker it blocked, on the
// opponent's turn: "at end of combat" is every combat.
func TestDreadWightParalyzesTheAttackerItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := dwCreature(g, me.ID, "Attacker", "", 2, 5)
	wight := dwCreature(g, opp.ID, "Dread Wight", oracleDreadWight, 3, 4)

	dwToEndOfCombat(t, g, opp.ID, attacker, wight)
	if dwCounters(t, g, attacker) != 1 || effectstest.HoldCount(t, g, attacker) != 1 {
		t.Fatalf("attacker: %d counters, %d holds; want one of each", dwCounters(t, g, attacker), effectstest.HoldCount(t, g, attacker))
	}
	if dwCounters(t, g, wight) != 0 {
		t.Error("the Wight paralyzed itself")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, attacker) {
		t.Error("the paralyzed attacker untapped")
	}
}

// A Wight in a combat it took no part in paralyzes nothing.
func TestDreadWightOutsideTheCombatDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwCreature(g, me.ID, "Dread Wight", oracleDreadWight, 3, 4)
	attacker := dwCreature(g, me.ID, "Attacker", "", 2, 5)
	blocker := dwCreature(g, opp.ID, "Wall", "", 0, 5)

	dwToEndOfCombat(t, g, opp.ID, attacker, blocker)
	if dwCounters(t, g, attacker) != 0 || dwCounters(t, g, blocker) != 0 {
		t.Error("the Wight paralyzed creatures it was not blocking or blocked by")
	}
}
