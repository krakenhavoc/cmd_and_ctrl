package testsex

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const foragingWickermawOracle = "11ceb2db-021e-422d-976e-7bfe3b67a8aa"

func TestForagingWickermawIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(foragingWickermawOracle)
	if !ok || spec.Name != "Foraging Wickermaw" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

func TestForagingWickermawSurveilsAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RfdCast(t, g, "Foraging Wickermaw", "Artifact Creature — Scarecrow", foragingWickermawOracle, 1, 3)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("entering did not surveil 1")
	}
}

// A colour named up front: the mana lands and the Wickermaw is that
// colour; a second activation this turn is refused, and the colour wears
// off at end of turn.
func TestForagingWickermawBecomesTheColorItMadeOnceATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	maw := effectstest.B12Push(g, me.ID, "Foraging Wickermaw", "Artifact Creature — Scarecrow", foragingWickermawOracle, 1, 3)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	if err := g.ActivateManaAbility(me.ID, maw, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the ability was activated with nothing to pay {1}")
	}
	if got := effectstest.EffectiveColors(t, g, maw); len(got) != 0 {
		t.Fatalf("an unpaid activation coloured the Wickermaw %v", got)
	}

	effectstest.FillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, maw, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{1}: add {G}: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"C": 1, "G": 1})
	if got := effectstest.EffectiveColors(t, g, maw); !slices.Equal(got, []string{"G"}) {
		t.Fatalf("Wickermaw colours = %v, want [G]", got)
	}
	if err := g.ActivateManaAbility(me.ID, maw, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("a second activation this turn was allowed")
	}

	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AangAdvanceToMain(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if got := effectstest.EffectiveColors(t, g, maw); len(got) != 0 {
		t.Errorf("after the turn the Wickermaw is %v, want colourless", got)
	}
}

// A colour picked after the activation: the Wickermaw becomes the picked
// colour when the pick is answered.
func TestForagingWickermawBecomesThePickedColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	maw := effectstest.B12Push(g, me.ID, "Foraging Wickermaw", "Artifact Creature — Scarecrow", foragingWickermawOracle, 1, 3)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateManaAbility(me.ID, maw, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick = %+v, want all five colours", pick)
	}
	if got := effectstest.EffectiveColors(t, g, maw); len(got) != 0 {
		t.Fatalf("the Wickermaw is %v before the colour is picked", got)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	if got := effectstest.EffectiveColors(t, g, maw); !slices.Equal(got, []string{"R"}) {
		t.Errorf("Wickermaw colours = %v, want [R]", got)
	}
}
