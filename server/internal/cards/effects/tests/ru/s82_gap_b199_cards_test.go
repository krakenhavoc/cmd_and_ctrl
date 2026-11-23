package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB199ForTheEmperorOracle  = "61dd58e0-9879-4e13-b12d-e7401d4cc42e"
	gapB199LaunchTheFleetOracle = "3362dec2-b157-4a00-8abf-0d7b5a14c6f8"
	gapB199BearType             = "Creature — Bear"
)

func TestGapB199EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB199ForTheEmperorOracle:  "For the Emperor!",
		gapB199LaunchTheFleetOracle: "Launch the Fleet",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB199ForTheEmperorPumpsOnlyYourCreaturesPresentAtResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", gapB199BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB199BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "For the Emperor!", "Sorcery", gapB199ForTheEmperorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 4 || tg != 4 {
		t.Errorf("my creature is %d/%d, want 4/4", p, tg)
	}
	for _, kw := range []string{"vigilance", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature did not gain %s", kw)
		}
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Errorf("an opposing creature was affected: power %d", p)
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", gapB199BearType, 2, 2)
	if p := effectstest.EffectivePower(t, g, late); p != 2 || effectstest.HasEffectiveKeyword(t, g, late, "vigilance") {
		t.Errorf("a creature that came later was affected: power %d", p)
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 2 || effectstest.HasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

func TestGapB199LaunchTheFleetStriveCostsOneMorePerExtraTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", gapB199BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", gapB199BearType, 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", gapB199BearType, 2, 2)
	for _, tc := range []struct {
		refs []game.TargetRef
		mv   int
	}{
		{effectstest.CardRefs(a), 1},
		{effectstest.CardRefs(a, b), 2},
		{effectstest.CardRefs(a, b, c), 3},
	} {
		if got := effectstest.SelfPriced(t, g, me, gapB199LaunchTheFleetOracle, "Sorcery", "{W}", game.ZoneHand, tc.refs).ManaValue(); got != tc.mv {
			t.Errorf("%d targets cost mana value %d, want %d", len(tc.refs), got, tc.mv)
		}
	}
}

func TestGapB199LaunchTheFleetGrantsTheSoldierTriggerToEachTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", gapB199BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", gapB199BearType, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", gapB199BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Launch the Fleet", "Sorcery", gapB199LaunchTheFleetOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	for _, id := range []uuid.UUID{a, b, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	// Two granted triggers go on the stack together: the controller orders them (CR 603.3b).
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Soldier"); got != 2 {
		t.Errorf("Soldier tokens after attacking with two granted creatures and one not: %d, want 2", got)
	}
}

func TestGapB199LaunchTheFleetWithNoTargetsDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", gapB199BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Launch the Fleet", "Sorcery", gapB199LaunchTheFleetOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB34AttackWith(t, g, opp.ID, a)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Soldier"); got != 0 {
		t.Errorf("Soldier tokens: %d, want 0", got)
	}
}
