package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const pupuUFOOracle = "b1412eda-e3a4-41b2-932e-795a0ba0f7f8"

func pushPuPuUFO(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "PuPu UFO", OracleID: pupuUFOOracle,
		TypeLine: "Artifact Creature — Construct Alien", ManaCost: "{2}",
		Power: 0, Toughness: 4, Owner: owner, Controller: owner,
	})
}

func pushTown(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Some Town", TypeLine: "Land — Town",
		Owner: owner, Controller: owner,
	})
}

func TestPuPuUFOHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	ufo := pushPuPuUFO(g, g.Seats[0].ID)
	if !hasString(effectiveAbilities(t, g, ufo), "flying") {
		t.Errorf("abilities = %v, want flying", effectiveAbilities(t, g, ufo))
	}
}

// "{T}: You may put a land card from your hand onto the battlefield."
// Only land cards are offered, declining is allowed, and the land is
// put rather than played — the turn's land drop is untouched.
func TestPuPuUFOPutsALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ufo := pushPuPuUFO(g, me.ID)
	land := handCardForTest(me, "Plains", "Basic Land — Plains", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")

	if err := g.ActivateCatalogAbility(me.ID, ufo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if !tappedOnBattlefield(t, g, ufo) {
		t.Error("the {T} cost was not paid")
	}
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 — \"you MAY put a land card\"", pick.ChooseMin, pick.ChooseMax)
	}
	if hasID(pick.ChooseCards, rock) {
		t.Error("a nonland card was offered")
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not reach the battlefield")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0: a put is not a play", n)
	}
}

// The {T} ability is a creature's tap ability, so a UFO that arrived
// this turn cannot use it (CR 302.6); the {3} ability has no {T} and
// can.
func TestPuPuUFOTapAbilityRespectsSummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ufo := pushPuPuUFO(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ufo {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	handCardForTest(me, "Plains", "Basic Land — Plains", "")
	if err := g.ActivateCatalogAbility(me.ID, ufo, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick UFO activated its {T} ability")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, ufo, 1, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("the {3} ability has no {T} and should not be gated: %v", err)
	}
}

// "{3}: Until end of turn, this creature's base power becomes equal to
// the number of Towns you control." Only YOUR Towns; the count is
// locked in on resolution (CR 608.2h); toughness is untouched; and the
// set ends with the turn.
func TestPuPuUFOBasePowerFromTowns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ufo := pushPuPuUFO(g, me.ID)
	pushTown(g, me.ID)
	pushTown(g, me.ID)
	pushTown(g, opp.ID)
	// A creature subtype that merely CONTAINS "Town" is not a Town.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Townsfolk", TypeLine: "Creature — Elephant Townsfolk",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, ufo, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := effectivePower(t, g, ufo); got != 0 {
		t.Fatalf("power = %d before the ability resolved, want 0", got)
	}
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, ufo); got != 2 {
		t.Errorf("power = %d, want 2 (your two Towns)", got)
	}
	if got := effectiveToughness(t, g, ufo); got != 4 {
		t.Errorf("toughness = %d, want the printed 4", got)
	}

	// A third Town after resolution does not move it.
	pushTown(g, me.ID)
	if got := effectivePower(t, g, ufo); got != 2 {
		t.Errorf("power = %d after a later Town, want 2 — the count is fixed on resolution", got)
	}

	advancePastCleanupForTest(t, g)
	if got := effectivePower(t, g, ufo); got != 0 {
		t.Errorf("power = %d after end of turn, want the printed 0", got)
	}
}

// A base-power SET is layer 7b: a +1/+1 counter still applies on top.
func TestPuPuUFOBasePowerKeepsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ufo := pushPuPuUFO(g, me.ID)
	pushTown(g, me.ID)
	if err := g.AddCounter(ufo, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, ufo, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, ufo); got != 1 {
		t.Errorf("layer power = %d, want 1 (one Town)", got)
	}
	var current int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(ufo); ok {
			current = c.CurrentPower()
		}
	})
	if current != 2 {
		t.Errorf("current power = %d, want 1 Town + 1 counter = 2", current)
	}
}
