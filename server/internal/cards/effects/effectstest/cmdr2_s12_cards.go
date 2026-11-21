package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S12RevealLandCase(t *testing.T, name, oracle string, matching, other string) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	match := HandCard(me, "Match", matching)
	nope := HandCard(me, "Nope", other)
	land := PlayLandFromHand(t, g, name, oracle)
	c := EntryRevealChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no reveal prompt")
	}
	if !HasID(c.ChooseCards, match) || HasID(c.ChooseCards, nope) {
		t.Errorf("candidates %v: want only the %s card", c.ChooseCards, matching)
	}
	AnswerEntryReveal(t, g, me.ID, match)
	if tapped, _ := BattlefieldCardTapped(g, land); tapped {
		t.Error("entered tapped after a reveal")
	}
	if !me.Hand.Contains(match) {
		t.Error("the revealed card left the hand")
	}
	// Declining enters tapped.
	g2 := NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	HandCard(me2, "Match", matching)
	land2 := PlayLandFromHand(t, g2, name, oracle)
	AnswerEntryReveal(t, g2, me2.ID)
	if tapped, _ := BattlefieldCardTapped(g2, land2); !tapped {
		t.Error("declining the reveal did not enter tapped")
	}
}

// s12Drive passes priority until the stack is empty and no prompt is left,
// answering trigger-order prompts as offered and pick_target prompts with
// `target` (the first illegal pick in `illegal` must be refused).
func S12Drive(t *testing.T, g *game.Game, chooser uuid.UUID, target uuid.UUID, illegal ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 40; i++ {
		answered := false
		for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
			if c == nil || c.Chooser != chooser {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoicePickTarget:
				for _, bad := range illegal {
					if err := g.ResolvePickTarget(c.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
						t.Errorf("an illegal target %s was accepted", bad)
					}
				}
				if err := g.ResolvePickTarget(c.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				answered = true
			}
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil && len(g.PendingChoices) == 0 {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
