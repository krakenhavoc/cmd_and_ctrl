package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// power_up_test.go — the enumerator half of #4527. A power-up ability
// this object has already activated is not a move (CR 702.193a,
// "Activate this ability only once"): the engine refuses it with
// ErrPowerUpUsed, so a bot seat must never be offered it, and both read
// game.Game.PowerUpUsed.

const oracleBraveBrawler = "a1a791be-71e6-4c1f-8717-08ca62d40de2"

func TestASpentPowerUpAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	brawler := battlefieldCard(g, active, game.Card{
		Name: "Brave Brawler", TypeLine: "Creature — Human Warrior Hero",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleBraveBrawler,
	})
	basics(g, active, "Plains", 5)
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler)
	if len(acts) != 1 {
		t.Fatalf("an unspent power-up ability: want 1 activation offered, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, brawler, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the power-up ability: %v", err)
	}
	basics(g, active, "Plains", 5)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 0 {
		t.Errorf("a spent power-up ability: want no activation offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(acts))
	}
}

const oracleWonderMan = "e2c49754-711e-43b3-82f1-5376a4d99701"

// #4571: under Wonder Man's "an additional time" the second activation
// is a move, the third is not, and the engine takes the same view.
func TestAPowerUpAbilityIsOfferedAgainUnderWonderMan(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	brawler := battlefieldCard(g, active, game.Card{
		Name: "Brave Brawler", TypeLine: "Creature — Human Warrior Hero",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleBraveBrawler,
	})
	battlefieldCard(g, active, game.Card{
		Name: "Wonder Man, Hollywood Hero", TypeLine: "Legendary Creature — Human Performer Hero",
		ManaCost: "{3}{R}{R}", Power: 4, Toughness: 4, OracleID: oracleWonderMan,
	})
	basics(g, active, "Plains", 5)
	advanceTo(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(active.ID, brawler, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	basics(g, active, "Plains", 5)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler)
	if len(acts) != 1 {
		t.Fatalf("after one activation under Wonder Man: want the second offered, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, brawler, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	basics(g, active, "Plains", 5)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 0 {
		t.Errorf("after two activations: want none offered, got %v", labels(acts))
	}
}

// #4571: mana restricted to power-up abilities (Quinjet Technician's
// {R}{R}) makes a power-up a move the pool could not otherwise pay, and
// the same pool is no use to the ability after it is spent.
func TestPowerUpOnlyManaMakesAPowerUpAbilityAMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	brawler := battlefieldCard(g, active, game.Card{
		Name: "Brave Brawler", TypeLine: "Creature — Human Warrior Hero",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleBraveBrawler,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	for _, c := range []string{"C", "C", "W"} {
		active.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 0 {
		t.Fatalf("{C}{C}{W} is short of {4}{W}: want no activation, got %v", labels(acts))
	}
	for range 2 {
		active.ManaPool.AddMana(game.ManaToken{Color: "R", Restrictions: []string{game.ManaRestrictPowerUp}})
	}
	acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler)
	if len(acts) != 1 {
		t.Fatalf("with {R}{R} for power-up abilities added: want 1 activation, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}

// #4571: while a power-up ban stands (Kang the Conqueror's extra turn)
// no power-up ability is a move.
func TestNoPowerUpAbilityIsOfferedWhileThePowerUpBanStands(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	brawler := battlefieldCard(g, active, game.Card{
		Name: "Brave Brawler", TypeLine: "Creature — Human Warrior Hero",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleBraveBrawler,
	})
	basics(g, active, "Plains", 5)
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 1 {
		t.Fatalf("before the ban: want 1 activation offered, got %v", labels(acts))
	}

	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, game.ScopeGame, uuid.Nil,
			[]game.Mod{{Kind: game.ModCantActivatePowerUp}}, g.UntilEndOfTurnDuration(), "Kang the Conqueror")
	})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 0 {
		t.Errorf("under the ban: want no activation offered, got %v — the engine refuses it", labels(acts))
	}
}
