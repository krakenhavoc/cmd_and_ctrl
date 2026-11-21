package teststo

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The owner's other case: casting Ugin himself with Vivi out. Ugin's cast
// trigger fills in the caster as its controller and nothing else, so with
// a target other than Vivi the pair goes on with no prompt.
func TestCastingUginWithViviOutDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vivi := pushColoredVivi(g, me.ID)
	victim := effectstest.PushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})

	effectstest.CastUginEyeOfTheStorms(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
		t.Fatalf("casting Ugin with Vivi out asked for an order: %+v", ch)
	}
	effectstest.SettleWithoutPrompts(t, g)
	if !effectstest.InExile(g, victim) {
		t.Error("Ugin's cast trigger exiles its target")
	}
	if got := effectstest.CounterCount(g, vivi, game.CounterPlusOne); got != 1 {
		t.Errorf("Vivi has %d +1/+1 counters, want 1", got)
	}
}

// Ugin's cast trigger aimed at Vivi still asks: a tie on one object.
func TestCastingUginTargetingViviAsks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vivi := pushColoredVivi(g, me.ID)

	effectstest.CastUginEyeOfTheStorms(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, vivi)
	if ch := effectstest.TriggerOrderPromptFor(g, me.ID); ch == nil {
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
		{"a controller fill-in", effectstest.BuiltPlainProbe, false},
		{"a Build that sets Params", effectstest.BuiltParamsProbe, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.PushCatalogPermanent(g, me.ID, "Probe", "Enchantment", c.oracle, false)
			effectstest.PushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", effectstest.FirebrandArcherOracle, false)
			effectstest.CastCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
			ch := effectstest.SettleUntilOrderPrompt(t, g)
			if c.ask && ch == nil {
				t.Error("want a trigger_order prompt")
			}
			if !c.ask && ch != nil {
				t.Errorf("want no prompt, got %+v", ch)
			}
		})
	}
}
