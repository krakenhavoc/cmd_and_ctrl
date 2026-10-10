package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_built_test.go — #2884 follow-up (owner answer
// (a), 2026-10-09): a row with a Build fill-in is read by its declared
// Effect's footprint, and its item is accepted only when the Build set
// nothing the footprint does not read.

const (
	builtPlainProbe  = "test-2884-built-plain-probe"
	builtParamsProbe = "test-2884-built-params-probe"
)

func init() {
	plain := WheneverYouCast(Noncreature(), "Built Plain Probe — you gain 1 life", Do(GainLife{Amount: 1}))
	plain.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Built Plain Probe — you gain 1 life")
		item.Controller, item.Owner = ev.Actor, ev.Actor
		return item
	}
	Register(Spec{OracleID: builtPlainProbe, Name: "Built Plain Probe", Triggered: []game.TriggeredAbility{plain}})

	params := WheneverYouCast(Noncreature(), "Built Params Probe — you gain 1 life", Do(GainLife{Amount: 1}))
	params.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Built Params Probe — you gain 1 life")
		item.Params.Amount = 3
		return item
	}
	Register(Spec{OracleID: builtParamsProbe, Name: "Built Params Probe", Triggered: []game.TriggeredAbility{params}})
}

// The owner's other case: casting Ugin himself with Vivi out. Ugin's cast
// trigger fills in the caster as its controller and nothing else, so with
// a target other than Vivi the pair goes on with no prompt.
func TestCastingUginWithViviOutDoesNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vivi := pushColoredVivi(g, me.ID)
	victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})

	castUginEyeOfTheStorms(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	if ch := triggerOrderPrompt(g); ch != nil {
		t.Fatalf("casting Ugin with Vivi out asked for an order: %+v", ch)
	}
	settleWithoutPrompts(t, g)
	if !inExile(g, victim) {
		t.Error("Ugin's cast trigger exiles its target")
	}
	if got := counterCount(g, vivi, game.CounterPlusOne); got != 1 {
		t.Errorf("Vivi has %d +1/+1 counters, want 1", got)
	}
}

// Ugin's cast trigger aimed at Vivi still asks: a tie on one object.
func TestCastingUginTargetingViviAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vivi := pushColoredVivi(g, me.ID)

	castUginEyeOfTheStorms(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, vivi)
	if ch := triggerOrderPromptFor(g, me.ID); ch == nil {
		t.Fatal("Ugin's cast trigger targeting Vivi must ask")
	}
}

// A Build that sets only the controller is accepted; one that records
// Params on the item is not, because the footprint does not read them.
func TestABuildIsAcceptedOnlyWhenItsItemIsPlain(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		ask          bool
	}{
		{"a controller fill-in", builtPlainProbe, false},
		{"a Build that sets Params", builtParamsProbe, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushCatalogPermanent(g, me.ID, "Probe", "Enchantment", c.oracle, false)
			pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
			castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
			ch := settleUntilOrderPrompt(t, g)
			if c.ask && ch == nil {
				t.Error("want a trigger_order prompt")
			}
			if !c.ask && ch != nil {
				t.Errorf("want no prompt, got %+v", ch)
			}
		})
	}
}
