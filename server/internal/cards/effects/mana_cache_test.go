package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_cache_test.go — Mana Cache (#1864, ADR 0106 §1 amendment
// 2026-10-16): an any-player MANA ability. The activator pays (CR
// 602.1a), the {C} goes into the activator's pool (CR 106.4), "their
// turn" is the activator's (CR 109.5), and the end-step trigger counts
// the lands of the player whose end step it is.

const manaCacheOracle = "1c09f2d1-413b-4a92-aa39-99c1a2106814"

func pushManaCache(g *game.Game, controller uuid.UUID, charge int) uuid.UUID {
	id := pushPermanentForTest(g, controller, "Mana Cache", manaCacheOracle, "Enchantment")
	if charge > 0 {
		findBattlefieldCardForTest(g, id).Counters = map[string]int{game.CounterCharge: charge}
	}
	return id
}

func chargeOn(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterCharge]
}

func TestManaCacheRowIsAnAnyPlayerManaAbility(t *testing.T) {
	rows := game.ManaAbilitiesForCard(game.Card{OracleID: manaCacheOracle, Name: "Mana Cache", TypeLine: "Enchantment"})
	if len(rows) != 1 {
		t.Fatalf("Mana Cache mana rows = %d, want 1", len(rows))
	}
	ab := rows[0]
	if !ab.AnyPlayer || ab.TapCost || ab.RemoveCounters == nil || ab.Produced != "{C}" {
		t.Errorf("row = %+v, want an any-player \"remove a charge counter: add {C}\"", ab)
	}
}

// TestManaCacheChargesForTheActivePlayersUntappedLands: each end step,
// one counter per untapped land of THAT player — the controller's own
// end step and another player's alike.
func TestManaCacheChargesForTheActivePlayersUntappedLands(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	cache := pushManaCache(g, a.ID, 0)
	for range 3 {
		pushLandFor(g, a.ID, "Mountain", "Basic Land — Mountain")
	}
	tapped := pushLandFor(g, a.ID, "Mountain", "Basic Land — Mountain")
	findBattlefieldCardForTest(g, tapped).Tapped = true
	for range 2 {
		pushLandFor(g, b.ID, "Forest", "Basic Land — Forest")
	}
	tappedB := pushLandFor(g, b.ID, "Forest", "Basic Land — Forest")

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := chargeOn(g, cache); got != 3 {
		t.Fatalf("after its controller's end step: %d charge counters, want 3 (three untapped lands, the tapped one not counted)", got)
	}

	// B's turn: B untaps (their tapped Forest is not B's untapped land
	// until then), then taps one Forest in the main phase.
	advanceToPrecombatMainOf(t, g, 1)
	findBattlefieldCardForTest(g, tappedB).Tapped = true
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := chargeOn(g, cache); got != 5 {
		t.Errorf("after another player's end step: %d charge counters, want 3+2", got)
	}
}

// TestManaCacheAnyPlayerMayActivateDuringTheirTurn is the activation:
// the player whose turn it is, controller or not, before the end step.
func TestManaCacheAnyPlayerMayActivateDuringTheirTurn(t *testing.T) {
	g := newCatalogGame(t)
	controller, active, other := g.Seats[0], g.Seats[1], g.Seats[2]
	cache := pushManaCache(g, controller.ID, 2)
	advanceToPrecombatMainOf(t, g, 1)

	// The controller and a bystander: not their turn.
	for _, p := range []*game.Player{controller, other} {
		err := g.ActivateManaAbility(p.ID, cache, 0, game.ManaAbilityParams{})
		if !errors.Is(err, game.ErrConditionNotMet) {
			t.Errorf("%s on another player's turn: err = %v, want ErrConditionNotMet", p.Name, err)
		}
	}
	if got := chargeOn(g, cache); got != 2 {
		t.Fatalf("a refused activation spent a counter: %d left", got)
	}

	// The active player reaches across: the counter comes off Mana Cache
	// and the {C} is theirs, not the controller's.
	if err := g.ActivateManaAbility(active.ID, cache, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the active player activates another player's Mana Cache: %v", err)
	}
	if got := chargeOn(g, cache); got != 1 {
		t.Errorf("charge counters %d, want 1", got)
	}
	if active.ManaPool.CountColor("C") != 1 {
		t.Errorf("activator's pool = %v, want one {C}", active.ManaPool)
	}
	if controller.ManaPool.Total() != 0 {
		t.Errorf("controller's pool = %v, want empty", controller.ManaPool)
	}
	if c := findBattlefieldCardForTest(g, cache); c.Controller != controller.ID {
		t.Errorf("Mana Cache changed controller to %v", c.Controller)
	}

	// Not in the end step, even on their own turn.
	advanceToEndStepOf(t, g, 1)
	if err := g.ActivateManaAbility(active.ID, cache, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("in the end step: err = %v, want ErrConditionNotMet", err)
	}
}

// TestManaCacheControllerActivatesOnTheirOwnTurn: the controller is one
// of "any player".
func TestManaCacheControllerActivatesOnTheirOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	controller, opp := g.Seats[0], g.Seats[1]
	cache := pushManaCache(g, controller.ID, 1)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(opp.ID, cache, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("an opponent on the controller's turn: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateManaAbility(controller.ID, cache, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("controller activates: %v", err)
	}
	if controller.ManaPool.CountColor("C") != 1 || chargeOn(g, cache) != 0 {
		t.Errorf("pool %v, counters %d; want one {C} and none left", controller.ManaPool, chargeOn(g, cache))
	}
	// With no counter left, the cost can't be paid.
	if err := g.ActivateManaAbility(controller.ID, cache, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activated with no charge counter left")
	}
}

// TestManaCacheIsNotAutoTapped: the planner never spends the counters,
// for anybody — the ability has no {T}, and the counters are a resource
// the activator chooses to spend. A cast it would pay for is not
// affordable without it.
func TestManaCacheIsNotAutoTapped(t *testing.T) {
	g := newCatalogGame(t)
	controller, active := g.Seats[0], g.Seats[1]
	cache := pushManaCache(g, controller.ID, 3)
	advanceToPrecombatMainOf(t, g, 1)
	one, err := game.ParseCost("{1}")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g.AutoTapPlanToppingUp(active.ID, one, 0, game.ManaSpendContext{}, nil, 0); ok {
		t.Errorf("the auto-tapper planned another player's Mana Cache")
	}
	if got := chargeOn(g, cache); got != 3 {
		t.Errorf("counters %d, want 3", got)
	}
}

// TestAnyPlayerManaAbilityRegistrationLimits pins the boot refusals.
func TestAnyPlayerManaAbilityRegistrationLimits(t *testing.T) {
	for name, ma := range map[string]ManaAbility{
		"tap":       {Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", AnyPlayer: true},
		"sacrifice": {Cost: ManaAbilityCost{Sacrifice: true}, Produced: "{C}", AnyPlayer: true},
		"exert":     {Cost: ManaAbilityCost{Exert: true}, Produced: "{C}", AnyPlayer: true},
		"hand":      {Cost: ManaAbilityCost{ExileSelf: true}, Zones: []game.ZoneKind{game.ZoneHand}, Produced: "{C}", AnyPlayer: true},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: an any-player mana ability was accepted", name)
				}
			}()
			checkAnyPlayerManaAbility("Test", 0, ma)
		}()
	}
	checkAnyPlayerManaAbility("Test", 0, ManaAbility{
		Cost:      ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
		Produced:  "{C}",
		AnyPlayer: true,
	})
}
