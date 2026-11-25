package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_6189_test.go — #6189: the cumulative upkeeps whose
// payment the prompt could not take, through the card layer.
//
//	702.24a ... Then you may pay [cost] for each age counter on it. If
//	        you don't, sacrifice it." If [cost] has choices associated
//	        with it, each choice is made separately for each age counter,
//	        ...
//
// "Cumulative upkeep {W} or {U}" (Jötun Owl Keeper, Krovikan Whispers,
// Arctic Nishoba, Earthen Goo) is one hybrid symbol per age counter,
// which is exactly that rule: the cost repeats as {W/U}{W/U} and each
// symbol is paid with either colour.

// cu6189Spec registers a test permanent with one cumulative upkeep and
// puts it onto the battlefield under seat 1.
func cu6189Spec(t *testing.T, g *game.Game, oracle, label string, ab game.TriggeredAbility) uuid.UUID {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: label, Triggered: []game.TriggeredAbility{ab}})
	me := g.Seats[1]
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: label, OracleID: oracle, TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
}

// A hybrid cumulative upkeep charges one hybrid symbol per age counter,
// and any mix of the two colours pays it.
func TestHybridCumulativeUpkeepTakesEitherColourPerAgeCounter(t *testing.T) {
	const oid = "test-6189-hybrid-upkeep"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	keeper := cu6189Spec(t, g, oid, "Test Owl Keeper",
		effects.CumulativeUpkeep("Test Owl Keeper — cumulative upkeep {W} or {U}", "{W/U}"))
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{W/U}" {
		t.Fatalf("first upkeep's cost = %q, want {W/U}", prompt.PayCost)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay with {U}: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, keeper); !ok {
		t.Fatal("a paid upkeep sacrificed the permanent")
	}

	// Second upkeep: two age counters, two hybrid symbols, paid with one
	// of each colour (CR 702.24a: each choice is made separately).
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{W/U}{W/U}" {
		t.Fatalf("second upkeep's cost = %q, want {W/U}{W/U}", prompt.PayCost)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "U"})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay with {W} and {U}: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, keeper); !ok {
		t.Error("a paid upkeep sacrificed the permanent")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want both symbols spent", me.ManaPool)
	}
}

// "Cumulative upkeep—Put a +1/+1 counter on a creature an opponent
// controls" (Sheltering Ancient) scales with the age counters, and
// paying puts them on the creatures named.
func TestCumulativeUpkeepCounterOnAnOpponentsCreature(t *testing.T) {
	const oid = "test-6189-counter-upkeep"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	ancient := cu6189Spec(t, g, oid, "Test Sheltering Ancient",
		effects.CumulativeUpkeepPaying("Test Sheltering Ancient — cumulative upkeep: put a +1/+1 counter on a creature an opponent controls",
			effects.CounterOnOpponentCreaturePayment(1, game.CounterPlusOne, "creature", "creatures", effects.QueryTypes("creature")),
			"creature", "creatures"))
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	action := prompt.PayAction()
	if action == nil || action.Kind != game.PayActionCounterOnOpponentCreature || action.Count != 1 {
		t.Fatalf("first upkeep's payment = %+v, want one counter on an opponent's creature", action)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("the opponent's creature's +1/+1 counters = %v, want 1", c)
	}
	if _, ok := effectstest.BattlefieldCard(g, ancient); !ok {
		t.Fatal("a paid upkeep sacrificed the permanent")
	}

	// Second upkeep: two counters, and the one creature may take both.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if a := prompt.PayAction(); a == nil || a.Count != 2 || !a.Repeats {
		t.Fatalf("second upkeep's payment = %+v, want two repeatable picks (CR 702.24a)", a)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{theirs, theirs}); err != nil {
		t.Fatalf("pay twice on one creature: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("the opponent's creature's +1/+1 counters = %v, want 3", c)
	}
}

// "Cumulative upkeep—Pay {B} and 1 life" (Infernal Darkness): both
// halves multiply with the age counters, and both are paid at once.
func TestCumulativeUpkeepManaAndLifeScalesBoth(t *testing.T) {
	const oid = "test-6189-mana-and-life-upkeep"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	cu6189Spec(t, g, oid, "Test Infernal Darkness",
		effects.CumulativeUpkeepPaying("Test Infernal Darkness — cumulative upkeep: pay {B} and 1 life",
			effects.ManaAndLifePayment("{B}", 1), "", ""))
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{B}" || prompt.LifeCost != 1 {
		t.Fatalf("first upkeep = %q plus %d life, want {B} plus 1", prompt.PayCost, prompt.LifeCost)
	}
	life := effectstest.LifeOf(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d → %d, want 1 paid with the mana", life, got)
	}
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{B}{B}" || prompt.LifeCost != 2 {
		t.Errorf("second upkeep = %q plus %d life, want {B}{B} plus 2", prompt.PayCost, prompt.LifeCost)
	}
}

// Snowfall's and Adarkar Unicorn's mana pays a cumulative upkeep (CR
// 106.6) and nothing else.
func TestUpkeepOnlyManaPaysTheCumulativeUpkeepPrompt(t *testing.T) {
	const oid = "test-6189-upkeep-only-mana"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	wall := cu6189Spec(t, g, oid, "Test Snowfall Wall",
		effects.CumulativeUpkeep("Test Snowfall Wall — cumulative upkeep {U}", "{U}"))
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "U", Restrictions: []string{game.ManaRestrictCumulativeUpkeep}})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay with upkeep-only mana: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, wall); !ok {
		t.Fatal("upkeep-only mana did not pay the cumulative upkeep")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want the restricted mana spent", me.ManaPool)
	}
}
