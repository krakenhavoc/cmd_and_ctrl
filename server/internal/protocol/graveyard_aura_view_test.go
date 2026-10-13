package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0154 decision 8 (#1780): no new wire field. While Animate Dead's
// enters trigger waits, the Aura's attached_to names the creature card in
// the opponent's graveyard, which is what lets a client show what is
// about to come back.
func TestAttachedToNamesTheGraveyardCardWhileTheTriggerWaits(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; g.Turn.Step != game.StepPrecombatMain; i++ {
		if i > 12 {
			t.Fatalf("never reached the main phase (at %s)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	zombie := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: zombie, Name: "Gravecrawler", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	aura := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: aura, Name: "Animate Dead", TypeLine: "Enchantment — Aura",
		OracleID: "c0d8fef4-65f4-4769-982d-b397d2b7e977", Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombie}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; g.Stack.Contains(aura); i++ {
		if i > 8 {
			t.Fatal("the Aura spell never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	v := ViewOfGame(g)
	var got *CardView
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == aura.String() {
			got = &v.Battlefield.Cards[i]
		}
	}
	if got == nil {
		t.Fatal("Animate Dead is not on the battlefield in the view")
	}
	if got.AttachedTo == nil || got.AttachedTo.Kind != string(game.TargetCard) || got.AttachedTo.ID != zombie.String() {
		t.Fatalf("attached_to = %+v, want the graveyard card %s", got.AttachedTo, zombie)
	}
	waiting := false
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == aura {
			waiting = true
		}
	}
	if !waiting {
		t.Error("the enters trigger is not waiting on the stack")
	}
}
