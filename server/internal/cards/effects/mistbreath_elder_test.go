package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mistbreathElderOracle = "bf9221f5-a4b2-4074-901c-0fa3c2bda80c"

func TestMistbreathElderIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(mistbreathElderOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Mistbreath Elder: %+v %v", spec, ok)
	}
}

func TestMistbreathElderBouncesAnotherCreatureAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elder := dskB1Creature(g, me.ID, "Mistbreath Elder", "Creature — Frog Warrior", mistbreathElderOracle, 2, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
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
	if _, ok := battlefieldCard(g, a); !ok {
		t.Error("the other creature left")
	}
	if got := counterOn(g, elder, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Elder: %d, want 1", got)
	}
}

func TestMistbreathElderAloneMayReturnItself(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		elder := dskB1Creature(g, me.ID, "Mistbreath Elder", "Creature — Frog Warrior", mistbreathElderOracle, 2, 2)
		advanceToStepOf(t, g, 0, game.StepUpkeep)
		passPriorityAroundTable(t, g)
		c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
		if c == nil {
			t.Fatal("no yes/no was asked")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		_, onField := battlefieldCard(g, elder)
		if yes && (onField || !me.Hand.Contains(elder)) {
			t.Errorf("yes: the Elder is on the battlefield=%v / in hand=%v, want in hand", onField, me.Hand.Contains(elder))
		}
		if !yes && !onField {
			t.Error("no: the Elder left the battlefield")
		}
		if got := counterOn(g, elder, game.CounterPlusOne); !yes && got != 0 {
			t.Errorf("counters %d, want 0", got)
		}
	}
}
