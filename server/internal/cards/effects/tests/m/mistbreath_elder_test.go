package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mistbreathElderOracle = "bf9221f5-a4b2-4074-901c-0fa3c2bda80c"

func TestMistbreathElderIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(mistbreathElderOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Mistbreath Elder: %+v %v", spec, ok)
	}
}

func TestMistbreathElderBouncesAnotherCreatureAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.DskB1Creature(g, me.ID, "Mistbreath Elder", "Creature — Frog Warrior", mistbreathElderOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 {
		t.Fatalf("pick %+v, want a mandatory choice between the two other creatures", pick)
	}
	for _, id := range pick.ChooseCards {
		if id == elder {
			t.Error("the Elder offered itself")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(b) {
		t.Error("the chosen creature is not in hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, a); !ok {
		t.Error("the other creature left")
	}
	if got := effectstest.CounterOn(g, elder, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Elder: %d, want 1", got)
	}
}

func TestMistbreathElderAloneMayReturnItself(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		elder := effectstest.DskB1Creature(g, me.ID, "Mistbreath Elder", "Creature — Frog Warrior", mistbreathElderOracle, 2, 2)
		effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
		if c == nil {
			t.Fatal("no yes/no was asked")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		_, onField := effectstest.BattlefieldCard(g, elder)
		if yes && (onField || !me.Hand.Contains(elder)) {
			t.Errorf("yes: the Elder is on the battlefield=%v / in hand=%v, want in hand", onField, me.Hand.Contains(elder))
		}
		if !yes && !onField {
			t.Error("no: the Elder left the battlefield")
		}
		if got := effectstest.CounterOn(g, elder, game.CounterPlusOne); !yes && got != 0 {
			t.Errorf("counters %d, want 0", got)
		}
	}
}
